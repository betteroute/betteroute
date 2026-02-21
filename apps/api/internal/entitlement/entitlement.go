// Package entitlement defines subscription tiers, feature gates, quota caps,
// and resolves the capability matrix for a workspace at request time.
//
// Plans live in application code (catalog.go). The database stores only
// billing state (which plan a workspace is on) and live usage counters.
package entitlement

import (
	"context"
	"fmt"

	"github.com/execrc/betteroute/internal/errs"
)

// Tier ranks subscription plans. Higher tiers inherit all lower-tier capabilities.
type Tier int

const (
	Free Tier = iota
	Pro
	Business
	Enterprise
)

type Feature int

const (
	// Free
	FeatureAPI Feature = iota
	FeatureLinkExpiration
	FeatureDeviceTargeting
	FeatureFolders
	FeatureTags
	FeatureCSVExport

	// Pro
	FeatureCustomDomains
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
	FeatureCustomOGMeta
	FeatureDeepLinking
	FeatureRealtimeAnalytics

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

// Check returns nil if the plan includes the feature, or a Forbidden error.
func (f Feature) Check(ctx Context) error {
	if ctx.CanAccess(f) {
		return nil
	}
	return errs.Forbidden(fmt.Sprintf(
		"%s is not available on the %s plan", f, ctx.Plan.Name,
	))
}

type Quota int

const (
	// Consumable — reset each usage cycle.
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

	// Configuration caps — read by other layers, not enforced by guard.
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

// Caps holds the numeric limits for all quotas within a plan.
type Caps [quotaCount]int

// Usage holds live consumption counters, indexed by Quota.
type Usage [quotaCount]int64

type Plan struct {
	Name string
	Tier Tier
	Caps Caps
}

// Overrides holds per-workspace customizations for enterprise deals.
type Overrides struct {
	Caps     *Caps            // nil = use plan defaults
	CapsSet  [quotaCount]bool // which quotas were explicitly set
	Features map[Feature]bool // extra features beyond the plan tier
}

// ParseOverrides unmarshals the JSONB columns from the subscription row.
func ParseOverrides(customLimits, customFeatures []byte) Overrides {
	var o Overrides
	if len(customLimits) > 0 {
		o.Caps, o.CapsSet = parseCustomCaps(customLimits)
	}
	if len(customFeatures) > 0 {
		o.Features = parseCustomFeatures(customFeatures)
	}
	return o
}

// Context is the resolved capability matrix for the current workspace request.
type Context struct {
	Plan           Plan
	usage          Usage
	customFeatures map[Feature]bool // enterprise overrides
}

// CanAccess returns true if the workspace plan (or override) permits the feature.
func (c Context) CanAccess(f Feature) bool {
	if c.customFeatures != nil && c.customFeatures[f] {
		return true
	}
	return c.Plan.Tier >= minTier[f]
}

const unlimited = -1

// CanCreate returns true if the workspace has enough quota capacity to create n items.
func (c Context) CanCreate(q Quota, n int) bool {
	if n == 0 {
		return true
	}
	if n < 0 {
		return false
	}
	limit := c.Plan.Caps[q]
	if limit == unlimited {
		return true
	}
	return c.usage[q]+int64(n) <= int64(limit)
}

// Used returns the current usage amount for a given quota.
func (c Context) Used(q Quota) int64 { return c.usage[q] }

// Resolve builds the entitlement Context for a workspace.
// Unknown plan IDs fall back to Free.
func Resolve(planID string, overrides Overrides, usage Usage) Context {
	p := lookup(planID)

	if overrides.Caps != nil {
		merged := p.Caps
		for q := range quotaCount {
			if overrides.CapsSet[q] {
				merged[q] = overrides.Caps[q]
			}
		}
		p.Caps = merged
	}

	return Context{
		Plan:           p,
		usage:          usage,
		customFeatures: overrides.Features,
	}
}

type contextKey struct{}

// NewContext attaches the entitlement context to the parent context.
func NewContext(parent context.Context, ent Context) context.Context {
	return context.WithValue(parent, contextKey{}, ent)
}

// FromContext extracts the entitlement Context. Returns a zero value if absent.
func FromContext(ctx context.Context) Context {
	c, _ := ctx.Value(contextKey{}).(Context)
	return c
}
