package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Service resolves short codes to redirect decisions.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new redirect service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: sqlc.New(pool)}
}

// Resolve looks up a short code, validates the link, and returns a Resolution.
func (s *Service) Resolve(ctx context.Context, code string) (*Resolution, error) {
	row, err := s.q.FindLinkByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if !row.IsActive {
		return nil, ErrInactive
	}

	now := time.Now()

	// Check scheduling: link hasn't started yet.
	if row.StartsAt != nil && now.Before(*row.StartsAt) {
		return nil, ErrNotStarted
	}

	// Check expiration: link has expired.
	if row.ExpiresAt != nil && now.After(*row.ExpiresAt) {
		if row.ExpirationUrl != nil && *row.ExpirationUrl != "" {
			return &Resolution{LinkID: row.ID, DestURL: *row.ExpirationUrl}, nil
		}
		return nil, ErrExpired
	}

	// Check click limit: link has reached max clicks.
	if row.MaxClicks != nil && row.ClickCount >= int64(*row.MaxClicks) {
		if row.ExpirationUrl != nil && *row.ExpirationUrl != "" {
			return &Resolution{LinkID: row.ID, DestURL: *row.ExpirationUrl}, nil
		}
		return nil, ErrClickLimitReached
	}

	// Increment click count async — doesn't block the redirect.
	// TODO: replace with batched writer when adding analytics pipeline.
	go func() {
		if err := s.q.IncrementClickCount(context.Background(), row.ID); err != nil {
			slog.Error("incrementing click count", "link_id", row.ID, "error", err)
		}
	}()

	return toResolution(row), nil
}

// toResolution maps a sqlc.Link to a redirect Resolution.
func toResolution(row sqlc.Link) *Resolution {
	dest := row.DestUrl

	// Append UTM parameters to destination URL if any are set.
	dest = appendUTM(dest, row)

	return &Resolution{
		LinkID:        row.ID,
		DestURL:       dest,
		OGTitle:       ptr.From(row.OgTitle),
		OGDescription: ptr.From(row.OgDescription),
		OGImage:       ptr.From(row.OgImage),
	}
}

// appendUTM appends stored UTM params to the destination URL query string.
func appendUTM(dest string, row sqlc.Link) string {
	params := []struct {
		key string
		val *string
	}{
		{"utm_source", row.UtmSource},
		{"utm_medium", row.UtmMedium},
		{"utm_campaign", row.UtmCampaign},
		{"utm_term", row.UtmTerm},
		{"utm_content", row.UtmContent},
	}

	// Fast path: skip URL parsing if no UTM params are set.
	hasAny := false
	for _, p := range params {
		if p.val != nil && *p.val != "" {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return dest
	}

	u, err := url.Parse(dest)
	if err != nil {
		return dest // invalid URL — return as-is
	}

	q := u.Query()
	for _, p := range params {
		if p.val != nil && *p.val != "" {
			q.Set(p.key, *p.val)
		}
	}
	u.RawQuery = q.Encode()

	return u.String()
}
