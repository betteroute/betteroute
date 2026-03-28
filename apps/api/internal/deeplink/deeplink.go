// Package deeplink manages deep linking: a read-only platform app catalog,
// workspace-owned app configurations, and per-link deep link resolution.
// Auto-detects apps from destination URLs and resolves platform-specific
// deep link URLs at link creation time.
package deeplink

import (
	"errors"
	"time"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/opt"
)

// PlatformApp is a pre-seeded app in the global catalog (YouTube, Spotify, etc.).
// Used for auto-detecting deep links from destination URLs.
type PlatformApp struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	IconURL        string    `json:"icon_url,omitempty"`
	URLPatterns    []string  `json:"url_patterns"`
	IOSScheme      string    `json:"ios_scheme,omitempty"`
	AndroidScheme  string    `json:"android_scheme,omitempty"`
	IOSAppID       string    `json:"ios_app_id,omitempty"`
	IOSBundleID    string    `json:"ios_bundle_id,omitempty"`
	IOSTeamID      string    `json:"ios_team_id,omitempty"`
	AndroidPackage string    `json:"android_package,omitempty"`
	AndroidSHA256  []string  `json:"android_sha256,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ResolvedLinks holds the deep link URLs resolved for a specific link.
// Stored in the deep_links table at link creation time.
type ResolvedLinks struct {
	PlatformAppID      string `json:"platform_app_id,omitempty"`
	WorkspaceAppID     string `json:"workspace_app_id,omitempty"`
	IOSDeepLink        string `json:"ios_deep_link,omitempty"`
	AndroidDeepLink    string `json:"android_deep_link,omitempty"`
	IOSFallbackURL     string `json:"ios_fallback_url,omitempty"`
	AndroidFallbackURL string `json:"android_fallback_url,omitempty"`
}

// WorkspaceApp is a user-managed iOS or Android app configuration.
// Used for AASA/assetlinks serving on custom domains.
type WorkspaceApp struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id"`
	CreatedBy          string    `json:"created_by,omitempty"`
	Name               string    `json:"name"`
	Platform           string    `json:"platform"` // "ios" or "android"
	BundleID           string    `json:"bundle_id,omitempty"`
	TeamID             string    `json:"team_id,omitempty"`
	AppStoreURL        string    `json:"app_store_url,omitempty"`
	PackageName        string    `json:"package_name,omitempty"`
	SHA256Fingerprints []string  `json:"sha256_fingerprints,omitempty"`
	PlayStoreURL       string    `json:"play_store_url,omitempty"`
	Scheme             string    `json:"scheme,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateWorkspaceAppInput is the input for creating a workspace app.
type CreateWorkspaceAppInput struct {
	Name               string   `json:"name"               validate:"required,min=1,max=100"`
	Platform           string   `json:"platform"            validate:"required,oneof=ios android"`
	BundleID           string   `json:"bundle_id"           validate:"omitempty,max=255"`
	TeamID             string   `json:"team_id"             validate:"omitempty,max=100"`
	AppStoreURL        string   `json:"app_store_url"       validate:"omitempty,url,max=2048"`
	PackageName        string   `json:"package_name"        validate:"omitempty,max=255"`
	SHA256Fingerprints []string `json:"sha256_fingerprints" validate:"omitempty"`
	PlayStoreURL       string   `json:"play_store_url"      validate:"omitempty,url,max=2048"`
	Scheme             string   `json:"scheme"              validate:"omitempty,max=500"`
}

// UpdateWorkspaceAppInput is the input for partially updating a workspace app.
type UpdateWorkspaceAppInput struct {
	Name               opt.Field[string]   `json:"name"                validate:"omitempty,min=1,max=100" swaggertype:"string"`
	BundleID           opt.Field[*string]  `json:"bundle_id"           validate:"omitempty,max=255" swaggertype:"string"`
	TeamID             opt.Field[*string]  `json:"team_id"             validate:"omitempty,max=100" swaggertype:"string"`
	AppStoreURL        opt.Field[*string]  `json:"app_store_url"       validate:"omitempty,url,max=2048" swaggertype:"string"`
	PackageName        opt.Field[*string]  `json:"package_name"        validate:"omitempty,max=255" swaggertype:"string"`
	SHA256Fingerprints opt.Field[[]string] `json:"sha256_fingerprints" validate:"omitempty" swaggertype:"array,string"`
	PlayStoreURL       opt.Field[*string]  `json:"play_store_url"      validate:"omitempty,url,max=2048" swaggertype:"string"`
	Scheme             opt.Field[*string]  `json:"scheme"              validate:"omitempty,max=500" swaggertype:"string"`
}

// Sentinel errors.
var ErrWorkspaceAppNotFound = errors.New("workspace app not found")

// ID generators
func newWorkspaceAppID() string { return "wsa_" + xid.New().String() }
