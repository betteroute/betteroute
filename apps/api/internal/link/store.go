package link

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/deeplink"
	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database operations for the link package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new link store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// Insert creates a new link record.
func (s *Store) Insert(ctx context.Context, l *Link) (*Link, error) {
	row, err := s.q.InsertLink(ctx, sqlc.InsertLinkParams{
		ID:            l.ID,
		WorkspaceID:   l.WorkspaceID,
		CreatedBy:     ptr.ToNonZero(l.CreatedBy),
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

// FindByID retrieves a single link by ID.
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

// List retrieves a paginated list of links for a workspace.
// Fetches limit+1 rows so the caller can determine has_more without a COUNT query.
func (s *Store) List(ctx context.Context, workspaceID string, limit, offset int) ([]Link, error) {
	rows, err := s.q.ListLinksByWorkspace(ctx, sqlc.ListLinksByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       ptr.ToInt32(limit + 1),
		Offset:      ptr.ToInt32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("listing links: %w", err)
	}

	links := make([]Link, len(rows))
	for i, row := range rows {
		links[i] = *toLink(row)
	}

	return links, nil
}

// Update partially updates a link.
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
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("updating link %s: %w", id, err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Link])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating link %s: %w", id, err)
	}
	return toLink(row), nil
}

// SoftDelete marks a link as deleted.
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

// UpsertDeepLink creates or updates deep link data for a link.
func (s *Store) UpsertDeepLink(ctx context.Context, linkID string, dl *deeplink.ResolvedLinks) error {
	_, err := s.q.UpsertDeepLink(ctx, sqlc.UpsertDeepLinkParams{
		LinkID:             linkID,
		PlatformAppID:      ptr.ToNonZero(dl.PlatformAppID),
		WorkspaceAppID:     ptr.ToNonZero(dl.WorkspaceAppID),
		IosDeepLink:        ptr.ToNonZero(dl.IOSDeepLink),
		AndroidDeepLink:    ptr.ToNonZero(dl.AndroidDeepLink),
		IosFallbackUrl:     ptr.ToNonZero(dl.IOSFallbackURL),
		AndroidFallbackUrl: ptr.ToNonZero(dl.AndroidFallbackURL),
	})
	if err != nil {
		return fmt.Errorf("upserting deep link for %s: %w", linkID, err)
	}
	return nil
}

func toLink(row sqlc.Link) *Link {
	return &Link{
		ID:               row.ID,
		WorkspaceID:      row.WorkspaceID,
		CreatedBy:        ptr.Val(row.CreatedBy),
		FolderID:         ptr.Val(row.FolderID),
		ShortCode:        row.ShortCode,
		DestURL:          row.DestUrl,
		Title:            ptr.Val(row.Title),
		Description:      ptr.Val(row.Description),
		IsActive:         row.IsActive,
		StartsAt:         row.StartsAt,
		ExpiresAt:        row.ExpiresAt,
		ExpirationURL:    ptr.Val(row.ExpirationUrl),
		MaxClicks:        row.MaxClicks,
		UTMSource:        ptr.Val(row.UtmSource),
		UTMMedium:        ptr.Val(row.UtmMedium),
		UTMCampaign:      ptr.Val(row.UtmCampaign),
		UTMTerm:          ptr.Val(row.UtmTerm),
		UTMContent:       ptr.Val(row.UtmContent),
		OGTitle:          ptr.Val(row.OgTitle),
		OGDescription:    ptr.Val(row.OgDescription),
		OGImage:          ptr.Val(row.OgImage),
		ClickCount:       row.ClickCount,
		UniqueClickCount: row.UniqueClickCount,
		LastClicked:      row.LastClickedAt,
		Notes:            ptr.Val(row.Notes),
		CreatedVia:       row.CreatedVia,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
