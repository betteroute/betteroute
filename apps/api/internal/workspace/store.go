package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database operations for the workspace package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new workspace store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// Insert creates a new workspace record.
func (s *Store) Insert(ctx context.Context, ws *Workspace) (*Workspace, error) {
	row, err := s.q.InsertWorkspace(ctx, sqlc.InsertWorkspaceParams{
		ID:     ws.ID,
		Name:   ws.Name,
		Slug:   ws.Slug,
		Status: ws.Status,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("inserting workspace: %w", err)
	}
	return toWorkspace(row), nil
}

// FindByID retrieves a single workspace by ID.
func (s *Store) FindByID(ctx context.Context, id string) (*Workspace, error) {
	row, err := s.q.FindWorkspaceByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying workspace %s: %w", id, err)
	}
	return toWorkspace(row), nil
}

// ListByUser retrieves all workspaces a user is a member of.
func (s *Store) ListByUser(ctx context.Context, userID string) ([]WithRole, error) {
	rows, err := s.q.ListWorkspacesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces for user %s: %w", userID, err)
	}
	out := make([]WithRole, len(rows))
	for i, row := range rows {
		out[i] = WithRole{
			Workspace: &Workspace{
				ID:        row.ID,
				Name:      row.Name,
				Slug:      row.Slug,
				Status:    row.Status,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			Role: rbac.Role(row.Role),
		}
	}
	return out, nil
}

// ListOwnedWorkspacePlans retrieves the subscription plan IDs for all active workspaces owned by the user.
func (s *Store) ListOwnedWorkspacePlans(ctx context.Context, userID string) ([]string, error) {
	plans, err := s.q.ListOwnedWorkspacePlans(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing owned plans for user %s: %w", userID, err)
	}
	return plans, nil
}

// UpdateStatus changes the status of a workspace and returns the updated record.
func (s *Store) UpdateStatus(ctx context.Context, id, status string) (*Workspace, error) {
	row, err := s.q.UpdateWorkspaceStatus(ctx, sqlc.UpdateWorkspaceStatusParams{
		ID:     id,
		Status: status,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating workspace status: %w", err)
	}
	return toWorkspace(row), nil
}

// Update partially updates a workspace.
func (s *Store) Update(ctx context.Context, id string, input UpdateInput) (*Workspace, error) {
	var u db.Update

	if input.Name.Set {
		u.Set("name", input.Name.Value)
	}
	if input.Slug.Set {
		u.Set("slug", input.Slug.Value)
	}

	if u.IsEmpty() {
		return s.FindByID(ctx, id)
	}

	sql, args := u.Build("workspaces", "id = ? AND deleted_at IS NULL", id)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Workspace])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if db.IsUniqueViolation(err) {
		return nil, ErrSlugTaken
	}
	if err != nil {
		return nil, fmt.Errorf("updating workspace %s: %w", id, err)
	}
	return toWorkspace(row), nil
}

// SoftDelete marks a workspace as deleted.
func (s *Store) SoftDelete(ctx context.Context, id string) error {
	if err := s.q.SoftDeleteWorkspace(ctx, id); err != nil {
		return fmt.Errorf("deleting workspace %s: %w", id, err)
	}
	return nil
}

// FindBySlugAndMember resolves a workspace by slug and verifies the user is a member.
// Returns the workspace and the member's role in one round-trip.
func (s *Store) FindBySlugAndMember(ctx context.Context, slug, userID string) (*Workspace, rbac.Role, error) {
	row, err := s.q.FindWorkspaceBySlugAndMember(ctx, sqlc.FindWorkspaceBySlugAndMemberParams{
		UserID: userID,
		Slug:   slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("resolving workspace %s for user %s: %w", slug, userID, err)
	}
	ws := &Workspace{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	return ws, rbac.Role(row.Role), nil
}

// InsertMember adds a user to a workspace with a specific role.
func (s *Store) InsertMember(ctx context.Context, workspaceID, userID string, role rbac.Role, invitedBy *string) error {
	err := s.q.InsertWorkspaceMember(ctx, sqlc.InsertWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        string(role),
		InvitedBy:   invitedBy,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return ErrAlreadyMember
		}
		return fmt.Errorf("inserting member: %w", err)
	}
	return nil
}

// FindMember retrieves a user's role within a workspace.
func (s *Store) FindMember(ctx context.Context, workspaceID, userID string) (*MemberRole, error) {
	row, err := s.q.FindWorkspaceMember(ctx, sqlc.FindWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotMember
	}
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}
	return &MemberRole{
		WorkspaceID: row.WorkspaceID,
		UserID:      row.UserID,
		Role:        rbac.Role(row.Role),
	}, nil
}

// ListMembers returns all active members of a workspace.
func (s *Store) ListMembers(ctx context.Context, workspaceID string) ([]*Member, error) {
	rows, err := s.q.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing members for workspace %s: %w", workspaceID, err)
	}
	out := make([]*Member, len(rows))
	for i, row := range rows {
		out[i] = &Member{
			UserID:    row.UserID,
			Name:      row.UserName,
			Email:     row.UserEmail,
			AvatarURL: ptr.From(row.UserAvatarUrl),
			Role:      rbac.Role(row.Role),
			JoinedAt:  row.CreatedAt,
		}
	}
	return out, nil
}

// UpdateMemberRole changes a user's role within a workspace.
func (s *Store) UpdateMemberRole(ctx context.Context, workspaceID, userID string, role rbac.Role) error {
	if err := s.q.UpdateWorkspaceMemberRole(ctx, sqlc.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        string(role),
	}); err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	return nil
}

// DeleteMember removes a user from a workspace.
func (s *Store) DeleteMember(ctx context.Context, workspaceID, userID string) error {
	if err := s.q.DeleteWorkspaceMember(ctx, sqlc.DeleteWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

// CountOwners returns the number of active owners in a workspace.
func (s *Store) CountOwners(ctx context.Context, workspaceID string) (int, error) {
	count, err := s.q.CountWorkspaceOwners(ctx, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("counting owners: %w", err)
	}
	return int(count), nil
}

// InsertInvitation creates a new pending workspace invitation.
func (s *Store) InsertInvitation(ctx context.Context, inv *Invitation, tokenHash string, invitedBy *string) (*Invitation, error) {
	row, err := s.q.InsertWorkspaceInvitation(ctx, sqlc.InsertWorkspaceInvitationParams{
		ID:          inv.ID,
		WorkspaceID: inv.WorkspaceID,
		Email:       inv.Email,
		Role:        string(inv.Role),
		TokenHash:   tokenHash,
		InvitedBy:   invitedBy,
		ExpiresAt:   inv.ExpiresAt,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrAlreadyInvited
		}
		return nil, fmt.Errorf("inserting invitation: %w", err)
	}
	return toInvitation(row), nil
}

// FindInvitationByToken retrieves an invitation using its plain-text token.
func (s *Store) FindInvitationByToken(ctx context.Context, plainToken string) (*Invitation, error) {
	row, err := s.q.FindWorkspaceInvitationByTokenHash(ctx, hashToken(plainToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("querying invitation by token: %w", err)
	}
	return toInvitation(row), nil
}

// ListInvitations returns all pending invitations for a workspace.
func (s *Store) ListInvitations(ctx context.Context, workspaceID string) ([]*Invitation, error) {
	rows, err := s.q.ListWorkspaceInvitations(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing invitations for workspace %s: %w", workspaceID, err)
	}
	out := make([]*Invitation, len(rows))
	for i, row := range rows {
		out[i] = toInvitation(row)
	}
	return out, nil
}

// AcceptInvitation marks a pending invitation as accepted.
func (s *Store) AcceptInvitation(ctx context.Context, invitationID string) error {
	if err := s.q.AcceptWorkspaceInvitation(ctx, invitationID); err != nil {
		return fmt.Errorf("accepting invitation %s: %w", invitationID, err)
	}
	return nil
}

// DeleteInvitation permanently removes a pending invitation.
func (s *Store) DeleteInvitation(ctx context.Context, invitationID, workspaceID string) error {
	if err := s.q.DeleteWorkspaceInvitation(ctx, sqlc.DeleteWorkspaceInvitationParams{
		ID:          invitationID,
		WorkspaceID: workspaceID,
	}); err != nil {
		return fmt.Errorf("deleting invitation %s: %w", invitationID, err)
	}
	return nil
}

func toWorkspace(row sqlc.Workspace) *Workspace {
	return &Workspace{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toInvitation(row sqlc.WorkspaceInvitation) *Invitation {
	return &Invitation{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Email:       row.Email,
		Role:        rbac.Role(row.Role),
		InvitedBy:   row.InvitedBy,
		ExpiresAt:   row.ExpiresAt,
		CreatedAt:   row.CreatedAt,
	}
}
