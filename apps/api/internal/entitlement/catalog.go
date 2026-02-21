package entitlement

import (
	"encoding/json"
	"fmt"
)

// minTier declares the minimum tier required to access each feature.
var minTier = [featureCount]Tier{
	// Free
	FeatureAPI:             Free,
	FeatureLinkExpiration:  Free,
	FeatureDeviceTargeting: Free,
	FeatureFolders:         Free,
	FeatureTags:            Free,
	FeatureCSVExport:       Free,

	// Pro
	FeatureCustomDomains:      Pro,
	FeaturePasswordProtection: Pro,
	FeatureClickExpiration:    Pro,
	FeatureOneTimeLinks:       Pro,
	FeatureUniqueVisitorLimit: Pro,
	FeatureLinkCloaking:       Pro,
	FeatureReferrerHiding:     Pro,
	FeatureExpirationURL:      Pro,
	FeatureLinkHealthMonitor:  Pro,
	FeatureCustomQRCode:       Pro,
	FeatureActivityLog:        Pro,
	FeatureJSONExport:         Pro,
	FeatureWebhooks:           Pro,
	FeatureLinkScheduling:     Pro,

	// Business
	FeatureFolderAccessControl: Business,
	FeatureGeoTargeting:        Business,
	FeatureBrowserTargeting:    Business,
	FeatureOSTargeting:         Business,
	FeatureLanguageTargeting:   Business,
	FeatureTimeRouting:         Business,
	FeatureDateRangeRouting:    Business,
	FeatureReferrerRestriction: Business,
	FeatureCountryBlocklist:    Business,
	FeatureEmailGate:           Business,
	FeatureABTesting:           Business,
	FeatureCustomOGMeta:        Business,
	FeatureDeepLinking:         Business,
	FeatureRealtimeAnalytics:   Business,

	// Enterprise
	FeatureSSO:        Enterprise,
	FeatureAuditLogs:  Enterprise,
	FeatureWhiteLabel: Enterprise,
	FeatureS3Export:   Enterprise,
}

// featureNames maps each Feature to its stable string key.
var featureNames = [featureCount]string{
	FeatureAPI:             "api_access",
	FeatureLinkExpiration:  "link_expiration",
	FeatureDeviceTargeting: "device_targeting",
	FeatureFolders:         "folders",
	FeatureTags:            "tags",
	FeatureCSVExport:       "csv_export",

	FeatureCustomDomains:      "custom_domains",
	FeaturePasswordProtection: "password_protection",
	FeatureClickExpiration:    "click_expiration",
	FeatureOneTimeLinks:       "one_time_links",
	FeatureUniqueVisitorLimit: "unique_visitor_limit",
	FeatureLinkCloaking:       "link_cloaking",
	FeatureReferrerHiding:     "referrer_hiding",
	FeatureExpirationURL:      "expiration_url",
	FeatureLinkHealthMonitor:  "link_health_monitor",
	FeatureCustomQRCode:       "custom_qr_code",
	FeatureActivityLog:        "activity_log",
	FeatureJSONExport:         "json_export",
	FeatureWebhooks:           "webhooks",
	FeatureLinkScheduling:     "link_scheduling",

	FeatureFolderAccessControl: "folder_access_control",
	FeatureGeoTargeting:        "geo_targeting",
	FeatureBrowserTargeting:    "browser_targeting",
	FeatureOSTargeting:         "os_targeting",
	FeatureLanguageTargeting:   "language_targeting",
	FeatureTimeRouting:         "time_routing",
	FeatureDateRangeRouting:    "date_range_routing",
	FeatureReferrerRestriction: "referrer_restriction",
	FeatureCountryBlocklist:    "country_blocklist",
	FeatureEmailGate:           "email_gate",
	FeatureABTesting:           "ab_testing",
	FeatureCustomOGMeta:        "custom_og_meta",
	FeatureDeepLinking:         "deep_linking",
	FeatureRealtimeAnalytics:   "realtime_analytics",

	FeatureSSO:        "sso_saml",
	FeatureAuditLogs:  "audit_logs",
	FeatureWhiteLabel: "white_label",
	FeatureS3Export:   "s3_export",
}

// quotaNames maps each Quota to its stable string key.
var quotaNames = [quotaCount]string{
	QuotaLinks:              "quota_links",
	QuotaClicks:             "quota_clicks",
	QuotaDomains:            "quota_domains",
	QuotaWebhooks:           "quota_webhooks",
	QuotaAPIKeys:            "quota_api_keys",
	QuotaMembers:            "quota_members",
	QuotaFolders:            "quota_folders",
	QuotaTags:               "quota_tags",
	QuotaWorkspaces:         "quota_workspaces",
	QuotaAPIRateLimit:       "quota_api_rate_limit",
	QuotaAnalyticsRetention: "quota_analytics_retention",
}

var (
	featureByName = buildIndex[Feature](featureNames[:])
	quotaByName   = buildIndex[Quota](quotaNames[:])
)

// catalog holds immutable plan definitions.
var catalog = map[string]Plan{
	"free": {
		Name: "Free", Tier: Free,
		Caps: Caps{
			QuotaLinks:              50,
			QuotaClicks:             1_000,
			QuotaDomains:            0,
			QuotaWebhooks:           0,
			QuotaAPIKeys:            1,
			QuotaMembers:            1,
			QuotaWorkspaces:         1,
			QuotaFolders:            20,
			QuotaTags:               unlimited,
			QuotaAPIRateLimit:       60,
			QuotaAnalyticsRetention: 30,
		},
	},
	"pro": {
		Name: "Pro", Tier: Pro,
		Caps: Caps{
			QuotaLinks:              5_000,
			QuotaClicks:             100_000,
			QuotaDomains:            10,
			QuotaWebhooks:           3,
			QuotaAPIKeys:            10,
			QuotaMembers:            5,
			QuotaWorkspaces:         3,
			QuotaFolders:            200,
			QuotaTags:               unlimited,
			QuotaAPIRateLimit:       600,
			QuotaAnalyticsRetention: 365,
		},
	},
	"business": {
		Name: "Business", Tier: Business,
		Caps: Caps{
			QuotaLinks:              50_000,
			QuotaClicks:             1_000_000,
			QuotaDomains:            100,
			QuotaWebhooks:           20,
			QuotaAPIKeys:            50,
			QuotaMembers:            25,
			QuotaWorkspaces:         10,
			QuotaFolders:            unlimited,
			QuotaTags:               unlimited,
			QuotaAPIRateLimit:       3_000,
			QuotaAnalyticsRetention: 1_095,
		},
	},
	"enterprise": {
		Name: "Enterprise", Tier: Enterprise,
		Caps: Caps{
			QuotaLinks:              unlimited,
			QuotaClicks:             unlimited,
			QuotaDomains:            unlimited,
			QuotaWebhooks:           unlimited,
			QuotaAPIKeys:            unlimited,
			QuotaMembers:            unlimited,
			QuotaWorkspaces:         unlimited,
			QuotaFolders:            unlimited,
			QuotaTags:               unlimited,
			QuotaAPIRateLimit:       unlimited,
			QuotaAnalyticsRetention: unlimited,
		},
	},
}

// lookup returns the Plan for the given ID, falling back to Free.
func lookup(id string) Plan {
	if p, ok := catalog[id]; ok {
		return p
	}
	return catalog["free"]
}

// buildIndex builds a name→enum reverse lookup and panics on missing entries.
func buildIndex[T ~int](names []string) map[string]T {
	idx := make(map[string]T, len(names))
	for i, name := range names {
		if name == "" {
			panic("entitlement: unnamed " + fmt.Sprintf("%T(%d)", T(0), i))
		}
		idx[name] = T(i)
	}
	return idx
}

// parseCustomCaps unmarshals {"quota_links": 100, ...} into a Caps array
// and a set of flags indicating which quotas were explicitly overridden.
func parseCustomCaps(raw []byte) (*Caps, [quotaCount]bool) {
	var m map[string]int
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return nil, [quotaCount]bool{}
	}
	var caps Caps
	var set [quotaCount]bool
	for k, v := range m {
		if q, ok := quotaByName[k]; ok {
			caps[q] = v
			set[q] = true
		}
	}
	return &caps, set
}

// parseCustomFeatures unmarshals {"deep_linking": true, ...} into a feature set.
func parseCustomFeatures(raw []byte) map[Feature]bool {
	var m map[string]bool
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return nil
	}
	out := make(map[Feature]bool, len(m))
	for k, v := range m {
		if f, ok := featureByName[k]; ok && v {
			out[f] = true
		}
	}
	return out
}
