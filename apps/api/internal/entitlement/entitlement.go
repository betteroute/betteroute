// Package entitlement defines subscription tiers, feature gates, quota limits,
// and resolves the capability matrix for a workspace at request time.
//
// Plans live in application code (catalog.go). The database stores only
// billing state (which plan a workspace is on) and live usage counters.
package entitlement

import "context"

// Unlimited signals that a quota has no cap.
const Unlimited = -1

// Tier ranks subscription plans. Higher tiers inherit all lower-tier capabilities.
type Tier int

const (
	Free Tier = iota
	Pro
	Business
	Enterprise
)

// Feature identifies a gated capability. Constants are grouped by minimum tier.
type Feature int

const (
	// Free
	FeatureAPI Feature = iota
	FeatureFolders
	FeatureTags
	FeatureCustomDomains

	// Pro
	FeatureCSVExport
	FeatureDeviceTargeting
	FeatureLinkExpiration
	FeaturePasswordProtection
	FeatureClickExpiration
	FeatureOneTimeLinks
	FeatureUniqueVisitorLimit
	FeatureLinkCloaking
	FeatureReferrerHiding
	FeatureExpirationURL
	FeatureLinkHealthMonitor
	FeatureCustomQRCode
	FeatureActivityLog
	FeatureJSONExport
	FeatureWebhooks
	FeatureLinkScheduling
	FeatureDeepLinking
	FeatureCustomOGMeta
	FeatureRealtimeAnalytics

	// Business
	FeatureFolderAccessControl
	FeatureGeoTargeting
	FeatureBrowserTargeting
	FeatureOSTargeting
	FeatureLanguageTargeting
	FeatureTimeRouting
	FeatureDateRangeRouting
	FeatureReferrerRestriction
	FeatureCountryBlocklist
	FeatureEmailGate
	FeatureABTesting

	// Enterprise
	FeatureSSO
	FeatureAuditLogs
	FeatureWhiteLabel
	FeatureS3Export

	featureCount // sentinel — must be last
)

// String returns the stable key for the feature (e.g. "deep_linking").
func (f Feature) String() string {
	if f >= 0 && int(f) < len(featureNames) {
		return featureNames[f]
	}
	return "unknown"
}

// Quota identifies a metered resource limit.
type Quota int

const (
	// Consumable — reset each billing cycle.
	QuotaLinks Quota = iota
	QuotaClicks

	// Allocated — persistent active counts, never reset.
	QuotaDomains
	QuotaWebhooks
	QuotaAPIKeys
	QuotaMembers
	QuotaFolders
	QuotaTags

	// Account-level — scoped to the user, not workspace.
	QuotaWorkspaces

	// Configuration — read by other layers, not enforced by guard.
	QuotaAPIRateLimit       // requests per minute
	QuotaAnalyticsRetention // retention in days

	quotaCount // sentinel — must be last
)

// String returns the stable key for the quota (e.g. "quota_links").
func (q Quota) String() string {
	if q >= 0 && int(q) < len(quotaNames) {
		return quotaNames[q]
	}
	return "unknown"
}

// Caps holds the numeric caps for all quotas within a plan.
// A value of Unlimited (-1) means no cap.
type Caps [quotaCount]int

// Usage holds live consumption counters, indexed by Quota.
type Usage [quotaCount]int64

// Plan is an immutable definition of a subscription tier's capabilities.
type Plan struct {
	Name string
	Tier Tier
	Caps Caps
}

// Context is the resolved capability matrix for the current workspace request.
type Context struct {
	Plan  Plan
	usage Usage
}

// HasFeature returns true if the workspace plan permits the feature.
func (c Context) HasFeature(f Feature) bool {
	return c.Plan.Tier >= featureTier[f]
}

// CanConsume returns true if the workspace has enough remaining capacity for n items.
func (c Context) CanConsume(q Quota, n int) bool {
	if n <= 0 {
		return n == 0
	}
	limit := c.Plan.Caps[q]
	if limit == Unlimited {
		return true
	}
	return c.usage[q]+int64(n) <= int64(limit)
}

// Used returns the current usage count for a given quota.
func (c Context) Used(q Quota) int64 { return c.usage[q] }

// Cap returns the plan cap for a given quota. Returns Unlimited (-1) for uncapped.
func (c Context) Cap(q Quota) int { return c.Plan.Caps[q] }

// Resolve builds the entitlement Context for a workspace.
// Unknown plan IDs fall back to Free.
func Resolve(planID string, usage Usage) Context {
	return Context{Plan: lookup(planID), usage: usage}
}

type contextKey struct{}

// resolver holds a lazy-evaluated entitlement Context.
// Resolved at most once per request, on first FromContext call.
// Safe without sync.Once because Fiber processes each request in a single goroutine.
type resolver struct {
	done   bool
	cached Context
	load   func(context.Context) Context
}

func (r *resolver) resolve(ctx context.Context) Context {
	if !r.done {
		r.cached = r.load(ctx)
		r.done = true
	}
	return r.cached
}

// WithResolver stores a lazy entitlement loader in context.
// The loader is called at most once, on the first FromContext access.
// Read-only endpoints that never call guard.Feature/guard.Quota pay zero DB cost.
func WithResolver(parent context.Context, load func(context.Context) Context) context.Context {
	return context.WithValue(parent, contextKey{}, &resolver{load: load})
}

// NewContext stores a pre-resolved entitlement context (useful for tests).
func NewContext(parent context.Context, ent Context) context.Context {
	return context.WithValue(parent, contextKey{}, &resolver{done: true, cached: ent})
}

// FromContext extracts the entitlement Context, triggering lazy resolution if needed.
// Returns a zero value (Free tier, all quotas blocked) if no resolver is set.
func FromContext(ctx context.Context) Context {
	r, ok := ctx.Value(contextKey{}).(*resolver)
	if !ok || r == nil {
		return Context{}
	}
	return r.resolve(ctx)
}
