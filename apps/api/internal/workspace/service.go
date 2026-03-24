package workspace

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/notify"
	"github.com/execrc/betteroute/internal/rbac"
)

const invitationTTL = 7 * 24 * time.Hour

// BillingProvisioner seeds subscription and usage rows for new workspaces.
type BillingProvisioner interface {
	Provision(ctx context.Context, workspaceID string) error
}

// Service implements workspace business logic.
type Service struct {
	store    *Store
	notifier notify.TeamNotifier
	webURL   string
	billing  BillingProvisioner
}

// NewService creates a new workspace service.
func NewService(store *Store, notifier notify.TeamNotifier, webURL string, billing BillingProvisioner) *Service {
	return &Service{store: store, notifier: notifier, webURL: webURL, billing: billing}
}

// ResolveAccess looks up a workspace by slug and verifies the user is a member.
func (s *Service) ResolveAccess(ctx context.Context, slug, userID string) (*Workspace, rbac.Role, error) {
	return s.store.FindBySlugAndMember(ctx, slug, userID)
}

// Create creates a new workspace, adds the creator as owner, and seeds billing state.
// The workspace is immediately active on the free plan. Upgrades happen via the
// billing checkout flow — no separate provisioning step required.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Workspace, error) {
	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = slugify(input.Name)
	}
	if slug == "" {
		return nil, ErrInvalidSlug
	}

	// Enforce workspace creation limits based on the user's highest plan.
	if err := s.checkWorkspaceLimit(ctx, userID); err != nil {
		return nil, err
	}

	ws, err := s.store.Insert(ctx, &Workspace{
		ID:     newWorkspaceID(),
		Name:   strings.TrimSpace(input.Name),
		Slug:   slug,
		Status: "active",
	})
	if err != nil {
		return nil, err
	}

	if err = s.store.InsertMember(ctx, ws.ID, userID, rbac.Owner, nil); err != nil {
		return nil, fmt.Errorf("adding owner: %w", err)
	}

	// Seed subscription + usage rows (free plan by default).
	if s.billing != nil {
		if err = s.billing.Provision(ctx, ws.ID); err != nil {
			slog.WarnContext(ctx, "seeding billing state", "error", err, "workspace_id", ws.ID)
		}
	}

	return ws, nil
}

// checkWorkspaceLimit verifies the user hasn't exceeded workspace creation limits.
// The limit is derived from the user's highest plan across all owned workspaces.
func (s *Service) checkWorkspaceLimit(ctx context.Context, userID string) error {
	plans, err := s.store.ListOwnedWorkspacePlans(ctx, userID)
	if err != nil {
		return err
	}

	maxAllowed := entitlement.Resolve("free", entitlement.Usage{}).Cap(entitlement.QuotaWorkspaces)
	for _, plan := range plans {
		limit := entitlement.Resolve(plan, entitlement.Usage{}).Cap(entitlement.QuotaWorkspaces)
		if limit == entitlement.Unlimited {
			return nil
		}
		if limit > maxAllowed {
			maxAllowed = limit
		}
	}

	if maxAllowed != -1 && len(plans) >= maxAllowed {
		return ErrLimitReached
	}
	return nil
}

// List returns all workspaces the user is a member of, with their role in each.
func (s *Service) List(ctx context.Context, userID string) ([]WithRole, error) {
	return s.store.ListByUser(ctx, userID)
}

// Get returns a workspace by ID.
func (s *Service) Get(ctx context.Context, workspaceID string) (*Workspace, error) {
	return s.store.FindByID(ctx, workspaceID)
}

// Update partially updates a workspace. Requires at least Admin role (enforced by middleware).
func (s *Service) Update(ctx context.Context, workspaceID string, input UpdateInput) (*Workspace, error) {
	if input.Name.Set {
		input.Name.Value = strings.TrimSpace(input.Name.Value)
	}
	if input.Slug.Set {
		input.Slug.Value = strings.TrimSpace(input.Slug.Value)
	}
	return s.store.Update(ctx, workspaceID, input)
}

// Delete soft-deletes a workspace. Requires Owner role (enforced by middleware).
func (s *Service) Delete(ctx context.Context, workspaceID string) error {
	return s.store.SoftDelete(ctx, workspaceID)
}

// ListMembers returns all members of a workspace.
func (s *Service) ListMembers(ctx context.Context, workspaceID string) ([]*Member, error) {
	return s.store.ListMembers(ctx, workspaceID)
}

// UpdateMember changes a member's role. Cannot demote the last owner.
func (s *Service) UpdateMember(ctx context.Context, workspaceID, targetUserID string, input UpdateMemberInput) error {
	member, err := s.store.FindMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}

	// Prevent demoting the last owner.
	if member.Role == rbac.Owner {
		owners, err := s.store.CountOwners(ctx, workspaceID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveOwner
		}
	}

	return s.store.UpdateMemberRole(ctx, workspaceID, targetUserID, input.Role)
}

// RemoveMember removes a user from a workspace. Cannot remove the last owner.
func (s *Service) RemoveMember(ctx context.Context, workspaceID, targetUserID string) error {
	member, err := s.store.FindMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}

	if member.Role == rbac.Owner {
		owners, err := s.store.CountOwners(ctx, workspaceID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveOwner
		}
	}

	return s.store.DeleteMember(ctx, workspaceID, targetUserID)
}

// Invite creates a pending invitation and sends the invite email.
func (s *Service) Invite(ctx context.Context, workspaceID, inviterID, inviterName, workspaceName string, input InviteInput) (*Invitation, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	plain, hash, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating invitation token: %w", err)
	}

	inv, err := s.store.InsertInvitation(ctx, &Invitation{
		ID:          newInvitationID(),
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        input.Role,
		ExpiresAt:   time.Now().Add(invitationTTL),
	}, hash, &inviterID)
	if err != nil {
		return nil, err
	}

	inviteURL := s.webURL + "/invitations/accept?token=" + plain
	go func() {
		if err := s.notifier.SendWorkspaceInviteEmail(context.Background(), email, inviterName, workspaceName, inviteURL); err != nil {
			slog.Error("sending workspace invite email", "error", err, "workspace_id", workspaceID)
		}
	}()

	return inv, nil
}

// ListInvitations returns pending invitations for a workspace.
func (s *Service) ListInvitations(ctx context.Context, workspaceID string) ([]*Invitation, error) {
	return s.store.ListInvitations(ctx, workspaceID)
}

// CancelInvitation deletes a pending invitation.
func (s *Service) CancelInvitation(ctx context.Context, workspaceID, invitationID string) error {
	return s.store.DeleteInvitation(ctx, invitationID, workspaceID)
}

// AcceptInvitation accepts a workspace invitation and adds the user as a member.
func (s *Service) AcceptInvitation(ctx context.Context, userID, userEmail string, input AcceptInvitationInput) (*WithRole, error) {
	inv, err := s.store.FindInvitationByToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(inv.Email, userEmail) {
		return nil, ErrInviteMismatch
	}

	if err = s.store.InsertMember(ctx, inv.WorkspaceID, userID, inv.Role, inv.InvitedBy); err != nil {
		return nil, err
	}

	if err = s.store.AcceptInvitation(ctx, inv.ID); err != nil {
		return nil, fmt.Errorf("marking invitation accepted: %w", err)
	}

	ws, err := s.store.FindByID(ctx, inv.WorkspaceID)
	if err != nil {
		return nil, err
	}

	return &WithRole{Workspace: ws, Role: inv.Role}, nil
}
