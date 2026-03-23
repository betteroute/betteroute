package entitlement

import "log/slog"

// featureTier declares the minimum tier required to access each feature.
// Grouped to mirror the Feature constant order.
var featureTier = [featureCount]Tier{
	// Free — explicitly listed for documentation; zero value is Free.
	FeatureAPI:           Free,
	FeatureFolders:       Free,
	FeatureTags:          Free,
	FeatureCustomDomains: Free,

	// Pro
	FeatureCSVExport:          Pro,
	FeatureDeviceTargeting:    Pro,
	FeatureLinkExpiration:     Pro,
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
	FeatureDeepLinking:        Pro,
	FeatureCustomOGMeta:       Pro,
	FeatureRealtimeAnalytics:  Pro,

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

	// Enterprise
	FeatureSSO:        Enterprise,
	FeatureAuditLogs:  Enterprise,
	FeatureWhiteLabel: Enterprise,
	FeatureS3Export:   Enterprise,
}

// featureNames maps each Feature to its stable string key.
var featureNames = [featureCount]string{
	FeatureAPI:           "api_access",
	FeatureFolders:       "folders",
	FeatureTags:          "tags",
	FeatureCustomDomains: "custom_domains",

	FeatureCSVExport:          "csv_export",
	FeatureDeviceTargeting:    "device_targeting",
	FeatureLinkExpiration:     "link_expiration",
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
	FeatureDeepLinking:        "deep_linking",
	FeatureCustomOGMeta:       "custom_og_meta",
	FeatureRealtimeAnalytics:  "realtime_analytics",

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

// catalog holds immutable plan definitions.
var catalog = map[string]Plan{
	"free": {
		Name: "Free", Tier: Free,
		Caps: Caps{
			QuotaLinks:              25,
			QuotaClicks:             1_000,
			QuotaDomains:            1,
			QuotaWebhooks:           0,
			QuotaAPIKeys:            1,
			QuotaMembers:            1,
			QuotaWorkspaces:         2,
			QuotaFolders:            5,
			QuotaTags:               10,
			QuotaAPIRateLimit:       60,
			QuotaAnalyticsRetention: 30,
		},
	},
	"pro": {
		Name: "Pro", Tier: Pro,
		Caps: Caps{
			QuotaLinks:              1_000,
			QuotaClicks:             50_000,
			QuotaDomains:            10,
			QuotaWebhooks:           5,
			QuotaAPIKeys:            10,
			QuotaMembers:            3,
			QuotaWorkspaces:         5,
			QuotaFolders:            100,
			QuotaTags:               Unlimited,
			QuotaAPIRateLimit:       600,
			QuotaAnalyticsRetention: 365,
		},
	},
	"business": {
		Name: "Business", Tier: Business,
		Caps: Caps{
			QuotaLinks:              10_000,
			QuotaClicks:             250_000,
			QuotaDomains:            50,
			QuotaWebhooks:           50,
			QuotaAPIKeys:            100,
			QuotaMembers:            25,
			QuotaWorkspaces:         50,
			QuotaFolders:            Unlimited,
			QuotaTags:               Unlimited,
			QuotaAPIRateLimit:       3_000,
			QuotaAnalyticsRetention: 1_095,
		},
	},
	"enterprise": {
		Name: "Enterprise", Tier: Enterprise,
		Caps: Caps{
			QuotaLinks:              Unlimited,
			QuotaClicks:             Unlimited,
			QuotaDomains:            Unlimited,
			QuotaWebhooks:           Unlimited,
			QuotaAPIKeys:            Unlimited,
			QuotaMembers:            Unlimited,
			QuotaWorkspaces:         Unlimited,
			QuotaFolders:            Unlimited,
			QuotaTags:               Unlimited,
			QuotaAPIRateLimit:       Unlimited,
			QuotaAnalyticsRetention: Unlimited,
		},
	},
	"selfhosted": {
		Name: "Self-Hosted", Tier: Enterprise,
		Caps: Caps{
			QuotaLinks:              Unlimited,
			QuotaClicks:             Unlimited,
			QuotaDomains:            Unlimited,
			QuotaWebhooks:           Unlimited,
			QuotaAPIKeys:            Unlimited,
			QuotaMembers:            Unlimited,
			QuotaWorkspaces:         Unlimited,
			QuotaFolders:            Unlimited,
			QuotaTags:               Unlimited,
			QuotaAPIRateLimit:       Unlimited,
			QuotaAnalyticsRetention: Unlimited,
		},
	},
}

// PlanName returns the display name for a plan ID, or "" if unknown.
func PlanName(id string) string {
	if p, ok := catalog[id]; ok {
		return p.Name
	}
	return ""
}

// lookup returns the Plan for the given ID, falling back to Free.
func lookup(id string) Plan {
	if p, ok := catalog[id]; ok {
		return p
	}
	slog.Warn("entitlement plan not found, falling back to free", "plan_id", id)
	return catalog["free"]
}
