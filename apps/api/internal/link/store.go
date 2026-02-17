package link

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// NullableFields tracks which nullable fields should be explicitly set (vs ignored).
type NullableFields struct {
	StartsAt  bool
	ExpiresAt bool
	MaxClicks bool
}

// Storer defines the interface for link storage operations.
type Storer interface {
	Insert(ctx context.Context, l *Link) (*Link, error)
	FindByID(ctx context.Context, id, workspaceID string) (*Link, error)
	List(ctx context.Context, workspaceID string, limit, offset int) ([]Link, int, error)
	Update(ctx context.Context, id, workspaceID string, input UpdateInput, nulls NullableFields) (*Link, error)
	SoftDelete(ctx context.Context, id, workspaceID string) error
}

// Store handles link database operations.
type Store struct {
	q *sqlc.Queries
}

// NewStore creates a new link store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Insert(ctx context.Context, l *Link) (*Link, error) {
	row, err := s.q.InsertLink(ctx, sqlc.InsertLinkParams{
		ID:            l.ID,
		WorkspaceID:   l.WorkspaceID,
		ShortCode:     l.ShortCode,
		DestUrl:       l.DestURL,
		Title:         ptr.ToNonZero(l.Title),
		Description:   ptr.ToNonZero(l.Description),
		StartsAt:      l.StartsAt,
		ExpiresAt:     l.ExpiresAt,
		ExpirationUrl: ptr.ToNonZero(l.ExpirationURL),
		MaxClicks:     l.MaxClicks,
		UtmSource:     ptr.ToNonZero(l.UTMSource),
		UtmMedium:     ptr.ToNonZero(l.UTMMedium),
		UtmCampaign:   ptr.ToNonZero(l.UTMCampaign),
		UtmTerm:       ptr.ToNonZero(l.UTMTerm),
		UtmContent:    ptr.ToNonZero(l.UTMContent),
		OgTitle:       ptr.ToNonZero(l.OGTitle),
		OgDescription: ptr.ToNonZero(l.OGDescription),
		OgImage:       ptr.ToNonZero(l.OGImage),
		Notes:         ptr.ToNonZero(l.Notes),
		CreatedVia:    l.CreatedVia,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrShortCodeTaken
		}
		return nil, fmt.Errorf("inserting link: %w", err)
	}
	return toLink(row), nil
}

func (s *Store) FindByID(ctx context.Context, id, workspaceID string) (*Link, error) {
	row, err := s.q.FindLinkByID(ctx, sqlc.FindLinkByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying link %s: %w", id, err)
	}
	return toLink(row), nil
}

func (s *Store) List(ctx context.Context, workspaceID string, limit, offset int) ([]Link, int, error) {
	rows, err := s.q.ListLinksByWorkspace(ctx, sqlc.ListLinksByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing links: %w", err)
	}

	total, err := s.q.CountLinksByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("counting links: %w", err)
	}

	links := make([]Link, len(rows))
	for i, row := range rows {
		links[i] = *toLink(row)
	}

	return links, int(total), nil
}

func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput, nulls NullableFields) (*Link, error) {
	row, err := s.q.UpdateLink(ctx, sqlc.UpdateLinkParams{
		ID:            id,
		WorkspaceID:   workspaceID,
		DestUrl:       input.DestURL,
		Title:         input.Title,
		Description:   input.Description,
		IsActive:      input.IsActive,
		SetStartsAt:   nulls.StartsAt,
		StartsAt:      input.StartsAt,
		SetExpiresAt:  nulls.ExpiresAt,
		ExpiresAt:     input.ExpiresAt,
		ExpirationUrl: input.ExpirationURL,
		SetMaxClicks:  nulls.MaxClicks,
		MaxClicks:     input.MaxClicks,
		UtmSource:     input.UTMSource,
		UtmMedium:     input.UTMMedium,
		UtmCampaign:   input.UTMCampaign,
		UtmTerm:       input.UTMTerm,
		UtmContent:    input.UTMContent,
		OgTitle:       input.OGTitle,
		OgDescription: input.OGDescription,
		OgImage:       input.OGImage,
		Notes:         input.Notes,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating link %s: %w", id, err)
	}
	return toLink(row), nil
}

func (s *Store) SoftDelete(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteLink(ctx, sqlc.SoftDeleteLinkParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting link %s: %w", id, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// toLink maps a sqlc.Link to a domain Link.
func toLink(row sqlc.Link) *Link {
	return &Link{
		ID:               row.ID,
		WorkspaceID:      row.WorkspaceID,
		ShortCode:        row.ShortCode,
		DestURL:          row.DestUrl,
		Title:            ptr.From(row.Title),
		Description:      ptr.From(row.Description),
		IsActive:         row.IsActive,
		StartsAt:         row.StartsAt,
		ExpiresAt:        row.ExpiresAt,
		ExpirationURL:    ptr.From(row.ExpirationUrl),
		MaxClicks:        row.MaxClicks,
		UTMSource:        ptr.From(row.UtmSource),
		UTMMedium:        ptr.From(row.UtmMedium),
		UTMCampaign:      ptr.From(row.UtmCampaign),
		UTMTerm:          ptr.From(row.UtmTerm),
		UTMContent:       ptr.From(row.UtmContent),
		OGTitle:          ptr.From(row.OgTitle),
		OGDescription:    ptr.From(row.OgDescription),
		OGImage:          ptr.From(row.OgImage),
		ClickCount:       row.ClickCount,
		UniqueClickCount: row.UniqueClickCount,
		LastClicked:      row.LastClickedAt,
		Notes:            ptr.From(row.Notes),
		CreatedVia:       row.CreatedVia,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
