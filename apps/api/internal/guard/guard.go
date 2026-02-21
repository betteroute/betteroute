// Package guard provides handler-level authorization checks.
//
//	func (h *Handler) Create(c fiber.Ctx) error {
//	    ctx := c.Context()
//	    if err := guard.Role(ctx, rbac.Member); err != nil { return err }
//	    if err := guard.Scope(ctx, rbac.ScopeLinksWrite); err != nil { return err }
//	    if err := guard.Feature(ctx, entitlement.FeatureCustomDomains); err != nil { return err }
//	    if err := guard.Quota(ctx, entitlement.QuotaLinks, len(inputs)); err != nil { return err }
//	    // ...
//	}
package guard

import (
	"context"
	"fmt"

	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/rbac"
)

// Role checks that the workspace member has at least the given role.
func Role(ctx context.Context, minRole rbac.Role) error {
	if !rbac.FromContext(ctx).Role.Has(minRole) {
		return errs.Forbidden("requires " + string(minRole) + " role or above")
	}
	return nil
}

// ScopeChecker is satisfied by *apikey.APIKey. Defined here so that guard
// does not import apikey (which imports guard via its handler).
type ScopeChecker interface {
	HasScope(rbac.Scope) bool
}

type scopeKey struct{}

// WithScope stores a scope checker in context.
// Called by the auth middleware when authenticating via API key.
func WithScope(ctx context.Context, sc ScopeChecker) context.Context {
	return context.WithValue(ctx, scopeKey{}, sc)
}

// Scope returns nil for session auth (no scope checker in context) —
// session users have full access without scope restrictions.
func Scope(ctx context.Context, s rbac.Scope) error {
	sc, _ := ctx.Value(scopeKey{}).(ScopeChecker)
	if sc == nil {
		return nil
	}
	if !sc.HasScope(s) {
		return errs.Forbidden("api key lacks required scope: " + string(s))
	}
	return nil
}

// Feature checks that the workspace plan includes the given feature.
func Feature(ctx context.Context, f entitlement.Feature) error {
	return f.Check(entitlement.FromContext(ctx))
}

// Quota checks that the workspace has enough remaining quota to create n units.
func Quota(ctx context.Context, q entitlement.Quota, n int) error {
	ent := entitlement.FromContext(ctx)
	if ent.CanCreate(q, n) {
		return nil
	}
	return errs.PaymentRequired(fmt.Sprintf(
		"%s quota exceeded (%d/%d) on the %s plan",
		q, ent.Used(q), ent.Plan.Caps[q], ent.Plan.Name,
	))
}
