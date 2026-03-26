// Package apikey handles programmatic API access via key-based authentication.
// Keys are workspace-scoped, SHA-256 hashed at rest, and shown once on creation.
//
// Permission model:
//   - all:        full access (capped at Member role — never Admin/Owner)
//   - read_only:  GET on all resources
//   - restricted: only specific scopes (e.g. "links:read", "tags:write")
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/execrc/betteroute/internal/rbac"
)

type Permission string

const (
	PermissionAll        Permission = "all"
	PermissionReadOnly   Permission = "read_only"
	PermissionRestricted Permission = "restricted"
)

// Valid reports whether the permission is a recognized value.
func (p Permission) Valid() bool {
	switch p {
	case PermissionAll, PermissionReadOnly, PermissionRestricted:
		return true
	}
	return false
}

func scopesToStrings(scopes []rbac.Scope) []string {
	if len(scopes) == 0 {
		return nil
	}
	out := make([]string, len(scopes))
	for i, s := range scopes {
		out[i] = string(s)
	}
	return out
}

func stringsToScopes(ss []string) []rbac.Scope {
	if len(ss) == 0 {
		return nil
	}
	scopes := make([]rbac.Scope, len(ss))
	for i, s := range ss {
		scopes[i] = rbac.Scope(s)
	}
	return scopes
}

// APIKey represents a workspace-scoped API key.
// The raw key is never stored — only the hash.
type APIKey struct {
	ID          string       `json:"id"`
	WorkspaceID string       `json:"workspace_id"`
	CreatedBy   string       `json:"created_by,omitempty"`
	Name        string       `json:"name"`
	KeyPrefix   string       `json:"key_prefix"`
	Permission  Permission   `json:"permission"`
	Scopes      []rbac.Scope `json:"scopes"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time   `json:"last_used_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// HasScope reports whether the key grants the requested scope.
//
//   - PermissionAll:        always true
//   - PermissionReadOnly:   true for any ":read" scope
//   - PermissionRestricted: true if explicitly listed, or if a write scope
//     for the same resource is listed and a read is requested (write implies read)
func (k *APIKey) HasScope(s rbac.Scope) bool {
	switch k.Permission {
	case PermissionAll:
		return true
	case PermissionReadOnly:
		return s.IsRead()
	}
	for _, granted := range k.Scopes {
		if granted == s {
			return true
		}
		if s.IsRead() && granted.IsWrite() && granted.Resource() == s.Resource() {
			return true
		}
	}
	return false
}

// CreateInput holds the fields required to create a new API key.
type CreateInput struct {
	Name       string       `json:"name"       validate:"required,min=1,max=100"`
	Permission Permission   `json:"permission" validate:"required"`
	Scopes     []rbac.Scope `json:"scopes"     validate:"omitempty"`
	ExpiresAt  *time.Time   `json:"expires_at" validate:"omitempty"`
}

// Creator holds the user fields returned by the JOIN query (FindByHashWithCreator).
// Avoids importing auth — the middleware maps this to auth.User.
type Creator struct {
	ID              string
	Name            string
	Email           string
	EmailVerifiedAt *time.Time
	AvatarURL       string
	Status          string
	OnboardedAt     *time.Time
	LastLoginAt     *time.Time
	Timezone        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

var (
	ErrNotFound       = errors.New("api key not found")
	ErrExpired        = errors.New("api key has expired")
	ErrInvalidScope   = errors.New("invalid scope")
	ErrScopesRequired = errors.New("restricted permission requires at least one scope")
)

const (
	// Prefix is the token prefix for API keys, shared with the auth middleware.
	Prefix    = "btr_"
	randomLen = 32
)

// generateKey produces a raw API key and its SHA-256 hash.
func generateKey() (plain, hash string, err error) {
	raw := make([]byte, randomLen)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generating api key: %w", err)
	}

	plain = Prefix + hex.EncodeToString(raw)
	hash = hashKey(plain)
	return plain, hash, nil
}

func hashKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

type contextKey struct{}

// NewContext returns a copy of parent with the API key attached.
func NewContext(parent context.Context, key *APIKey) context.Context {
	return context.WithValue(parent, contextKey{}, key)
}

// FromContext extracts the API key from ctx, returns nil when the request uses session auth.
func FromContext(ctx context.Context) *APIKey {
	key, _ := ctx.Value(contextKey{}).(*APIKey)
	return key
}
