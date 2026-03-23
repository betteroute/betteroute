package middleware

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"

	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/sqlc"
)

type entitlementStore interface {
	FindEntitlement(ctx context.Context, workspaceID string) (sqlc.FindEntitlementRow, error)
	RolloverUsageCycle(ctx context.Context, workspaceID string) (int64, error)
}

// Entitlement stores a lazy resolver for the workspace's capability matrix.
// The actual DB query is deferred until a guard.Feature or guard.Quota check
// Must run after Workspace middleware.
func Entitlement(store entitlementStore) fiber.Handler {
	return func(c fiber.Ctx) error {
		workspaceID := rbac.FromContext(c.Context()).WorkspaceID

		c.SetContext(entitlement.WithResolver(c.Context(), func(ctx context.Context) entitlement.Context {
			return loadEntitlement(ctx, store, workspaceID)
		}))
		return c.Next()
	}
}

// loadEntitlement fetches the plan and usage counters from the database.
//
// Rollover strategy:
//   - Paid plans: Polar webhook is the primary rollover (ForceResetWorkspaceUsage).
//     The check here is a safety net for delayed or lost webhooks.
//   - Free plans: No webhook exists. This is the only rollover mechanism.
//
// On DB error, returns a zero Context (Free tier, all quotas blocked) as a safe default.
func loadEntitlement(ctx context.Context, store entitlementStore, workspaceID string) entitlement.Context {
	ent, err := store.FindEntitlement(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return entitlement.Resolve("free", entitlement.Usage{})
	}
	if err != nil {
		slog.ErrorContext(ctx, "entitlement query failed, denying access", "error", err, "workspace_id", workspaceID)
		return entitlement.Context{}
	}

	// Rollover only when the billing cycle has actually expired.
	cycleExpired := ent.UsagePeriodEnd != nil && ent.UsagePeriodEnd.Before(time.Now())
	if cycleExpired {
		if _, err := store.RolloverUsageCycle(ctx, workspaceID); err != nil {
			slog.WarnContext(ctx, "usage cycle rollover failed", "error", err, "workspace_id", workspaceID)
		}
	}

	// Consumables reset to zero on rollover; allocated counters survive cycle resets.
	var usage entitlement.Usage
	if !cycleExpired {
		usage[entitlement.QuotaLinks] = int64(ent.LinksUsage)
		usage[entitlement.QuotaClicks] = ent.ClicksUsage
	}
	usage[entitlement.QuotaDomains] = int64(ent.DomainsActive)
	usage[entitlement.QuotaWebhooks] = int64(ent.WebhooksActive)
	usage[entitlement.QuotaAPIKeys] = int64(ent.ApiKeysActive)
	usage[entitlement.QuotaMembers] = int64(ent.MembersActive)
	usage[entitlement.QuotaFolders] = int64(ent.FoldersActive)
	usage[entitlement.QuotaTags] = int64(ent.TagsActive)

	return entitlement.Resolve(ent.PlanID, usage)
}
