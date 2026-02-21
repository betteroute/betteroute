// Package rbac provides role-based and scope-based access control for
// workspace resources. Roles govern what a human can do; scopes govern
// what an API key is allowed to do.
package rbac

import (
	"context"
	"strings"
)

// Role is a user's permission level within a workspace.
type Role string

const (
	// Owner role has full access including destructive workspace actions.
	Owner  Role = "owner"
	Admin  Role = "admin"
	Member Role = "member"
	Viewer Role = "viewer"
)

func (r Role) rank() int {
	switch r {
	case Owner:
		return 4
	case Admin:
		return 3
	case Member:
		return 2
	case Viewer:
		return 1
	default:
		return 0
	}
}

// Has reports whether the role meets or exceeds the minimum required role.
func (r Role) Has(minRole Role) bool {
	return r.rank() >= minRole.rank()
}

// Context holds the workspace ID and role for the current request.
type Context struct {
	WorkspaceID string
	Role        Role
}

type contextKey struct{}

// NewContext returns a copy of parent with the rbac Context attached.
func NewContext(parent context.Context, rctx Context) context.Context {
	return context.WithValue(parent, contextKey{}, rctx)
}

// FromContext extracts the rbac Context from ctx.
// Returns a zero Context if called outside a workspace-scoped route.
func FromContext(ctx context.Context) Context {
	rctx, _ := ctx.Value(contextKey{}).(Context)
	return rctx
}

// Scope is a fine-grained permission for API key access.
type Scope string

const (
	ScopeLinksRead     Scope = "links:read"
	ScopeLinksWrite    Scope = "links:write"
	ScopeFoldersRead   Scope = "folders:read"
	ScopeFoldersWrite  Scope = "folders:write"
	ScopeTagsRead      Scope = "tags:read"
	ScopeTagsWrite     Scope = "tags:write"
	ScopeDomainsRead   Scope = "domains:read"
	ScopeDomainsWrite  Scope = "domains:write"
	ScopeWebhooksRead  Scope = "webhooks:read"
	ScopeWebhooksWrite Scope = "webhooks:write"
	ScopeAnalyticsRead Scope = "analytics:read"
	ScopeWorkspaceRead Scope = "workspace:read"
)

// Valid reports whether s is a recognised scope string.
func (s Scope) Valid() bool {
	switch s {
	case ScopeLinksRead, ScopeLinksWrite,
		ScopeFoldersRead, ScopeFoldersWrite,
		ScopeTagsRead, ScopeTagsWrite,
		ScopeDomainsRead, ScopeDomainsWrite,
		ScopeWebhooksRead, ScopeWebhooksWrite,
		ScopeAnalyticsRead, ScopeWorkspaceRead:
		return true
	}
	return false
}

// Resource returns the resource portion of the scope (e.g. "links").
func (s Scope) Resource() string {
	if i := strings.IndexByte(string(s), ':'); i >= 0 {
		return string(s)[:i]
	}
	return string(s)
}

// IsRead reports whether the scope is a read scope.
func (s Scope) IsRead() bool { return strings.HasSuffix(string(s), ":read") }

// IsWrite reports whether the scope is a write scope.
func (s Scope) IsWrite() bool { return strings.HasSuffix(string(s), ":write") }
