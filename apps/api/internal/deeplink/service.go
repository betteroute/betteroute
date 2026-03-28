package deeplink

import (
	"context"
	"net/url"
	"strings"
)

// Service handles platform app detection, workspace app CRUD,
// and per-link deep link resolution.
type Service struct {
	store *Store
}

// NewService creates a new deeplink service.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// DetectApp looks up a platform app from a destination URL.
// Returns nil if no known app matches the URL hostname.
func (s *Service) DetectApp(ctx context.Context, destURL string) (*PlatformApp, error) {
	hostname := extractHostname(destURL)
	if hostname == "" {
		return nil, nil
	}
	return s.store.FindByHostname(ctx, hostname)
}

// ResolveDeepLinks auto-detects an app from destURL and builds deep link URLs.
// Returns nil if no known app matches.
func (s *Service) ResolveDeepLinks(ctx context.Context, destURL string) (*ResolvedLinks, error) {
	app, err := s.DetectApp(ctx, destURL)
	if err != nil || app == nil {
		return nil, err
	}

	pathAndQuery := extractPathAndQuery(destURL)

	dl := &ResolvedLinks{PlatformAppID: app.ID}

	if app.IOSScheme != "" {
		dl.IOSDeepLink = resolveScheme(app.IOSScheme, pathAndQuery)
	}
	if app.AndroidScheme != "" {
		dl.AndroidDeepLink = resolveScheme(app.AndroidScheme, pathAndQuery)
	}

	// Fallback URLs: app store pages for users without the app installed.
	if app.IOSAppID != "" {
		dl.IOSFallbackURL = "https://apps.apple.com/app/id" + app.IOSAppID
	} else {
		dl.IOSFallbackURL = destURL
	}
	if app.AndroidPackage != "" {
		dl.AndroidFallbackURL = "https://play.google.com/store/apps/details?id=" + app.AndroidPackage
	} else {
		dl.AndroidFallbackURL = destURL
	}

	return dl, nil
}

// ListPlatformApps returns all platform apps.
func (s *Service) ListPlatformApps(ctx context.Context) ([]PlatformApp, error) {
	return s.store.List(ctx)
}

// GetPlatformApp retrieves a platform app by ID.
func (s *Service) GetPlatformApp(ctx context.Context, id string) (*PlatformApp, error) {
	return s.store.FindByID(ctx, id)
}

// CreateWorkspaceApp creates a new workspace app.
func (s *Service) CreateWorkspaceApp(ctx context.Context, workspaceID, userID string, input CreateWorkspaceAppInput) (*WorkspaceApp, error) {
	wa := &WorkspaceApp{
		ID:                 newWorkspaceAppID(),
		WorkspaceID:        workspaceID,
		CreatedBy:          userID,
		Name:               input.Name,
		Platform:           input.Platform,
		BundleID:           input.BundleID,
		TeamID:             input.TeamID,
		AppStoreURL:        input.AppStoreURL,
		PackageName:        input.PackageName,
		SHA256Fingerprints: input.SHA256Fingerprints,
		PlayStoreURL:       input.PlayStoreURL,
		Scheme:             input.Scheme,
	}
	return s.store.InsertWorkspaceApp(ctx, wa)
}

// GetWorkspaceApp retrieves a workspace app by ID.
func (s *Service) GetWorkspaceApp(ctx context.Context, id, workspaceID string) (*WorkspaceApp, error) {
	return s.store.FindWorkspaceAppByID(ctx, id, workspaceID)
}

// ListWorkspaceApps returns all workspace apps for a workspace.
func (s *Service) ListWorkspaceApps(ctx context.Context, workspaceID string) ([]WorkspaceApp, error) {
	return s.store.ListWorkspaceApps(ctx, workspaceID)
}

// UpdateWorkspaceApp partially updates a workspace app.
func (s *Service) UpdateWorkspaceApp(ctx context.Context, id, workspaceID string, input UpdateWorkspaceAppInput) (*WorkspaceApp, error) {
	return s.store.UpdateWorkspaceApp(ctx, id, workspaceID, input)
}

// DeleteWorkspaceApp soft-deletes a workspace app.
func (s *Service) DeleteWorkspaceApp(ctx context.Context, id, workspaceID string) error {
	return s.store.SoftDeleteWorkspaceApp(ctx, id, workspaceID)
}

// extractHostname returns the hostname from a URL, stripping "www." prefix.
func extractHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}

// extractPathAndQuery returns the path, query, and fragment from a URL.
func extractPathAndQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	result := u.Path
	if u.RawQuery != "" {
		result += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		result += "#" + u.Fragment
	}
	return strings.TrimPrefix(result, "/")
}

// resolveScheme replaces {path} in the scheme template with the actual path.
func resolveScheme(scheme, pathAndQuery string) string {
	return strings.ReplaceAll(scheme, "{path}", pathAndQuery)
}
