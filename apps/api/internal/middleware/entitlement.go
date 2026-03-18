package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"

	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/sqlc"
)

type entitlementQuerier interface {
	FindEntitlement(ctx context.Context, workspaceID string) (sqlc.FindEntitlementRow, error)
	RolloverUsageCycle(ctx context.Context, workspaceID string) (int64, error)
}

// Entitlement resolves the capability matrix for the current workspace.
// Must run after Workspace middleware. No subscription defaults to Free.
func Entitlement(q entitlementQuerier) fiber.Handler {
	return func(c fiber.Ctx) error {
		workspaceID := rbac.FromContext(c.Context()).WorkspaceID

		// Idempotent constraint (<= NOW()) prevents Postgres lock contention unless expired.
		_, _ = q.RolloverUsageCycle(c.Context(), workspaceID)

		row, err := q.FindEntitlement(c.Context(), workspaceID)
		if errors.Is(err, pgx.ErrNoRows) {
			ent := entitlement.Resolve("free", entitlement.Usage{})
			c.SetContext(entitlement.NewContext(c.Context(), ent))
			return c.Next()
		}
		if err != nil {
			return errs.Internal("").WithCause(err)
		}

		var usage entitlement.Usage
		usage[entitlement.QuotaLinks] = int64(row.LinksUsage)
		usage[entitlement.QuotaClicks] = row.ClicksUsage
		usage[entitlement.QuotaDomains] = int64(row.DomainsActive)
		usage[entitlement.QuotaWebhooks] = int64(row.WebhooksActive)
		usage[entitlement.QuotaAPIKeys] = int64(row.ApiKeysActive)
		usage[entitlement.QuotaMembers] = int64(row.MembersActive)
		usage[entitlement.QuotaFolders] = int64(row.FoldersActive)
		usage[entitlement.QuotaTags] = int64(row.TagsActive)

		ent := entitlement.Resolve(row.PlanID, usage)
		c.SetContext(entitlement.NewContext(c.Context(), ent))
		return c.Next()
	}
}
