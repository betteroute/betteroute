package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/apikey"
	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/workspace"
)

// Workspace resolves the workspace slug from the URL and verifies the
// authenticated user is a member. Injects workspace ID and role into
// the request's context.Context.
//
// When the request is authenticated via API key, the role is capped:
//   - The key's workspace_id must match the resolved workspace.
//   - The effective role is min(membership role, Member) — API keys never
//     get Admin or Owner privileges.
func Workspace(svc *workspace.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		slug := c.Params("slug")
		ctx := c.Context()
		user := auth.FromContext(ctx).User

		ws, role, err := svc.ResolveAccess(ctx, slug, user.ID)
		if err != nil {
			return errs.NotFound("workspace", "")
		}

		// API key path: verify workspace match and cap role.
		if key := apikey.FromContext(ctx); key != nil {
			if key.WorkspaceID != ws.ID {
				return errs.NotFound("workspace", "")
			}
			if role.Has(rbac.Admin) {
				role = rbac.Member
			}
		}

		// Suspended workspaces can only access billing routes to resolve payment issues.
		if ws.Status == "suspended" {
			if !strings.Contains(c.Path(), "/billing") {
				return errs.Forbidden("workspace suspended")
			}
		}

		// Store workspace in context so handlers can read it without re-querying.
		ctx = workspace.NewContext(ctx, ws)

		c.SetContext(rbac.NewContext(ctx, rbac.Context{
			WorkspaceID: ws.ID,
			Status:      ws.Status,
			Role:        role,
		}))

		return c.Next()
	}
}
