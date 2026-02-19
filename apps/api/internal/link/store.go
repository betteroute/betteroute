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

// Store handles link database operations.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new link store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

func (s *Store) Insert(ctx context.Context, l *Link) (*Link, error) {
	row, err := s.q.InsertLink(ctx, sqlc.InsertLinkParams{
		ID:            l.ID,
		WorkspaceID:   l.WorkspaceID,
		FolderID:      ptr.ToNonZero(l.FolderID),
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

func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Link, error) {
	var u db.Update

	if input.DestURL.Set {
		u.Set("dest_url", input.DestURL.Value)
	}
	if input.IsActive.Set {
		u.Set("is_active", input.IsActive.Value)
	}
	if input.Title.Set {
		u.Set("title", input.Title.Value)
	}
	if input.Description.Set {
		u.Set("description", input.Description.Value)
	}
	if input.StartsAt.Set {
		u.Set("starts_at", input.StartsAt.Value)
	}
	if input.ExpiresAt.Set {
		u.Set("expires_at", input.ExpiresAt.Value)
	}
	if input.ExpirationURL.Set {
		u.Set("expiration_url", input.ExpirationURL.Value)
	}
	if input.MaxClicks.Set {
		u.Set("max_clicks", input.MaxClicks.Value)
	}
	if input.UTMSource.Set {
		u.Set("utm_source", input.UTMSource.Value)
	}
	if input.UTMMedium.Set {
		u.Set("utm_medium", input.UTMMedium.Value)
	}
	if input.UTMCampaign.Set {
		u.Set("utm_campaign", input.UTMCampaign.Value)
	}
	if input.UTMTerm.Set {
		u.Set("utm_term", input.UTMTerm.Value)
	}
	if input.UTMContent.Set {
		u.Set("utm_content", input.UTMContent.Value)
	}
	if input.OGTitle.Set {
		u.Set("og_title", input.OGTitle.Value)
	}
	if input.OGDescription.Set {
		u.Set("og_description", input.OGDescription.Value)
	}
	if input.OGImage.Set {
		u.Set("og_image", input.OGImage.Value)
	}
	if input.Notes.Set {
		u.Set("notes", input.Notes.Value)
	}
	if input.FolderID.Set {
		u.Set("folder_id", input.FolderID.Value)
	}

	if u.IsEmpty() {
		return s.FindByID(ctx, id, workspaceID)
	}

	sql, args := u.Build("links", "id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Link])
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
		FolderID:         ptr.From(row.FolderID),
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
