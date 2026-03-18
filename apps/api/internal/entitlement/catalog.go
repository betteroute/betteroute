package entitlement

import "log/slog"

// minTier declares the minimum tier required to access each feature.
var minTier = [featureCount]Tier{
	// Free
	FeatureFolders:       Free,
	FeatureTags:          Free,
	FeatureCustomDomains: Free,
	FeatureAPI:           Free,

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

	// Team
	FeatureFolderAccessControl: Team,
	FeatureGeoTargeting:        Team,
	FeatureBrowserTargeting:    Team,
	FeatureOSTargeting:         Team,
	FeatureLanguageTargeting:   Team,
	FeatureTimeRouting:         Team,
	FeatureDateRangeRouting:    Team,
	FeatureReferrerRestriction: Team,
	FeatureCountryBlocklist:    Team,
	FeatureEmailGate:           Team,
	FeatureABTesting:           Team,
	FeatureCustomOGMeta:        Team,
	FeatureDeepLinking:         Team,
	FeatureRealtimeAnalytics:   Team,

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

// catalog holds immutable plan definitions.
var catalog = map[string]Plan{
	"free": {
		Name: "Free", Tier: Free,
		Caps: Caps{
			QuotaLinks:              25,
			QuotaClicks:             1_000,
			QuotaDomains:            3,
			QuotaWebhooks:           0,
			QuotaAPIKeys:            1,
			QuotaMembers:            3,
			QuotaWorkspaces:         3,
			QuotaFolders:            10,
			QuotaTags:               25,
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
			QuotaMembers:            10,
			QuotaWorkspaces:         10,
			QuotaFolders:            100,
			QuotaTags:               unlimited,
			QuotaAPIRateLimit:       600,
			QuotaAnalyticsRetention: 365,
		},
	},
	"team": {
		Name: "Team", Tier: Team,
		Caps: Caps{
			QuotaLinks:              10_000,
			QuotaClicks:             250_000,
			QuotaDomains:            50,
			QuotaWebhooks:           50,
			QuotaAPIKeys:            100,
			QuotaMembers:            25,
			QuotaWorkspaces:         50,
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
	"selfhosted": {
		Name: "Self-Hosted", Tier: Enterprise,
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
