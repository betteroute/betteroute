package link

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/xid"
)

// primaryDomain is the default domain for short links.
// TODO: replace with DB-driven domains when the domains feature is built.
var primaryDomain = "http://localhost:8080"

// Service handles link business logic.
type Service struct {
	store *Store
}

// NewService creates a new link service.
func NewService(store *Store) *Service {
	return &Service{store: store}
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
					return s.enrichShortURL(created), nil
				}
				if !errors.Is(err, ErrShortCodeTaken) {
					return nil, err
				}
			}
		}
		return nil, err
	}

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
func (s *Service) List(ctx context.Context, workspaceID string, limit, offset int) ([]Link, int, error) {
	links, total, err := s.store.List(ctx, workspaceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for i := range links {
		s.enrichShortURL(&links[i])
	}
	return links, total, nil
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
	return s.store.SoftDelete(ctx, id, workspaceID)
}

// enrichShortURL computes the full short URL for a link.
// TODO: use l.Domain (per-link) when custom domains are added.
func (s *Service) enrichShortURL(l *Link) *Link {
	l.ShortURL = primaryDomain + "/" + l.ShortCode
	return l
}
