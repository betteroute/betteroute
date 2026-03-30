package link

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/deeplink"
	"github.com/execrc/betteroute/internal/usage"
)

// primaryDomain is the default domain for short links.
// TODO: replace with DB-driven domains when the domains feature is built.
var primaryDomain = "http://localhost:8080"

// Service handles link business logic.
type Service struct {
	store  *Store
	appSvc *deeplink.Service
	meter  *usage.Meter
}

// NewService creates a new link service.
func NewService(store *Store, appSvc *deeplink.Service, meter *usage.Meter) *Service {
	return &Service{store: store, appSvc: appSvc, meter: meter}
}

// Create generates a short code and persists a new link.
func (s *Service) Create(ctx context.Context, workspaceID, userID, createdVia string, input CreateInput) (*Link, error) {
	code := input.ShortCode

	if code == "" {
		var err error
		code, err = generateShortCode()
		if err != nil {
			return nil, fmt.Errorf("generating short code: %w", err)
		}
	}

	if input.WorkspaceAppID != "" && s.appSvc != nil {
		if _, err := s.appSvc.GetWorkspaceApp(ctx, input.WorkspaceAppID, workspaceID); err != nil {
			return nil, fmt.Errorf("invalid workspace_app_id: %w", err)
		}
	}

	l := &Link{
		ID:            "lnk_" + xid.New().String(),
		WorkspaceID:   workspaceID,
		CreatedBy:     userID,
		FolderID:      input.FolderID,
		ShortCode:     code,
		DestURL:       input.DestURL,
		Title:         input.Title,
		Description:   input.Description,
		StartsAt:      input.StartsAt,
		ExpiresAt:     input.ExpiresAt,
		ExpirationURL: input.ExpirationURL,
		MaxClicks:     input.MaxClicks,
		UTMSource:     input.UTMSource,
		UTMMedium:     input.UTMMedium,
		UTMCampaign:   input.UTMCampaign,
		UTMTerm:       input.UTMTerm,
		UTMContent:    input.UTMContent,
		OGTitle:       input.OGTitle,
		OGDescription: input.OGDescription,
		OGImage:       input.OGImage,
		Notes:         input.Notes,
		CreatedVia:    createdVia,
	}

	created, err := s.store.Insert(ctx, l)
	if err != nil {
		// Retry with new code on collision (only if auto-generated)
		if errors.Is(err, ErrShortCodeTaken) && input.ShortCode == "" {
			for range maxRetries {
				code, err = generateShortCode()
				if err != nil {
					return nil, fmt.Errorf("generating short code: %w", err)
				}
				l.ShortCode = code
				created, err = s.store.Insert(ctx, l)
				if err == nil {
					s.trackLinkCreated(ctx, created, input.WorkspaceAppID)
					return s.enrichShortURL(created), nil
				}
				if !errors.Is(err, ErrShortCodeTaken) {
					return nil, err
				}
			}
		}
		return nil, err
	}

	s.trackLinkCreated(ctx, created, input.WorkspaceAppID)
	return s.enrichShortURL(created), nil
}

// Get retrieves a link by ID within a workspace.
func (s *Service) Get(ctx context.Context, id, workspaceID string) (*Link, error) {
	l, err := s.store.FindByID(ctx, id, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.enrichShortURL(l), nil
}

// List returns paginated links for a workspace.
// Returns limit+1 rows so the handler can detect has_more.
func (s *Service) List(ctx context.Context, workspaceID string, limit, offset int) ([]Link, error) {
	links, err := s.store.List(ctx, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range links {
		s.enrichShortURL(&links[i])
	}
	return links, nil
}

// Update partially updates a link.
func (s *Service) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Link, error) {
	l, err := s.store.Update(ctx, id, workspaceID, input)
	if err != nil {
		return nil, err
	}
	return s.enrichShortURL(l), nil
}

// Delete soft-deletes a link.
func (s *Service) Delete(ctx context.Context, id, workspaceID string) error {
	if err := s.store.SoftDelete(ctx, id, workspaceID); err != nil {
		return err
	}

	if err := s.meter.Adjust(ctx, workspaceID, usage.Links, -1); err != nil {
		slog.WarnContext(ctx, "adjusting link usage", "error", err, "workspace_id", workspaceID)
	}

	return nil
}

// enrichShortURL computes the full short URL for a link.
// TODO: use l.Domain (per-link) when custom domains are added.
func (s *Service) enrichShortURL(l *Link) *Link {
	l.ShortURL = primaryDomain + "/" + l.ShortCode
	return l
}

// trackLinkCreated handles post-creation side effects: usage tracking and deep link detection.
func (s *Service) trackLinkCreated(ctx context.Context, l *Link, workspaceAppID string) {
	if err := s.meter.Adjust(ctx, l.WorkspaceID, usage.Links, 1); err != nil {
		slog.WarnContext(ctx, "adjusting link usage", "error", err, "workspace_id", l.WorkspaceID)
	}
	s.detectAndSaveDeepLinks(ctx, l, workspaceAppID)
}

// detectAndSaveDeepLinks auto-detects a platform app from the destination URL
// and persists the resolved deep link URLs, natively applying Custom Workspace Apps.
func (s *Service) detectAndSaveDeepLinks(ctx context.Context, l *Link, workspaceAppID string) {
	if s.appSvc == nil && workspaceAppID == "" {
		return
	}

	var dl *deeplink.ResolvedLinks
	var err error

	if s.appSvc != nil {
		dl, err = s.appSvc.ResolveDeepLinks(ctx, l.DestURL)
		if err != nil {
			slog.WarnContext(ctx, "detecting deep link", "link_id", l.ID, "error", err)
		}
	}

	if dl == nil {
		if workspaceAppID == "" {
			return // no platform app and no custom app, strictly nothing to save
		}
		dl = &deeplink.ResolvedLinks{}
	}

	if workspaceAppID != "" {
		dl.WorkspaceAppID = workspaceAppID
	}

	if err := s.store.UpsertDeepLink(ctx, l.ID, dl); err != nil {
		slog.WarnContext(ctx, "upserting deep link", "link_id", l.ID, "error", err)
	}
}
