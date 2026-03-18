// Package workspace manages workspaces — the multi-tenancy boundary for all
// resources. Every link, folder, tag, and API key belongs to one workspace.
package workspace

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/opt"
	"github.com/execrc/betteroute/internal/rbac"
)

// Domain types.

// Workspace is the multi-tenancy boundary for all resources.
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WithRole wraps a workspace with the current user's role.
// Used in API responses so the frontend knows what the viewer can do.
type WithRole struct {
	*Workspace
	Role rbac.Role `json:"role"`
}

// Member is a workspace member with their user identity.
type Member struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      rbac.Role `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// Invitation is a pending workspace invitation.
type Invitation struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Email       string    `json:"email"`
	Role        rbac.Role `json:"role"`
	InvitedBy   *string   `json:"invited_by,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemberRole represents a user's role within a workspace.
type MemberRole struct {
	WorkspaceID string
	UserID      string
	Role        rbac.Role
}

// CreateInput is the payload for creating a new workspace.
type CreateInput struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Slug string `json:"slug" validate:"omitempty,min=1,max=50"`
}

// UpdateInput is the payload for partially updating a workspace.
type UpdateInput struct {
	Name opt.Field[string] `json:"name" validate:"omitempty,min=1,max=100" swaggertype:"string"`
	Slug opt.Field[string] `json:"slug" validate:"omitempty,min=1,max=50"  swaggertype:"string"`
}

// InviteInput is the payload for sending a workspace invitation.
type InviteInput struct {
	Email string    `json:"email" validate:"required,email,max=254"`
	Role  rbac.Role `json:"role"  validate:"required,oneof=admin member viewer"`
}

// UpdateMemberInput is the payload for changing a member's role.
type UpdateMemberInput struct {
	Role rbac.Role `json:"role" validate:"required,oneof=admin member viewer"`
}

// AcceptInvitationInput is the payload for accepting a workspace invitation.
type AcceptInvitationInput struct {
	Token string `json:"token" validate:"required"`
}

var (
	ErrNotFound          = errors.New("workspace not found")
	ErrSlugTaken         = errors.New("slug already in use")
	ErrNotMember         = errors.New("not a workspace member")
	ErrAlreadyMember     = errors.New("user is already a workspace member")
	ErrCannotRemoveOwner = errors.New("cannot remove the last owner")
	ErrInvalidSlug       = errors.New("workspace name produced an invalid slug")
	ErrTokenInvalid      = errors.New("invitation token is invalid or expired")
	ErrInviteMismatch    = errors.New("invitation is for a different email address")
	ErrAlreadyInvited    = errors.New("a pending invitation already exists for this email")
	ErrLimitReached      = errors.New("workspace limit reached")
)

// ID generators.

func newWorkspaceID() string  { return "ws_" + xid.New().String() }
func newInvitationID() string { return "inv_" + xid.New().String() }

// Token helpers.

func generateToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	hash = hashToken(plain)
	return
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// Slug helpers.

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a workspace name into a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = strings.TrimRight(s[:50], "-")
	}
	return s
}
