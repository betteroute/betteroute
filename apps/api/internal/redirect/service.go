package redirect

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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

	if row.ExpiresAt != nil && row.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpired
	}

	return toResolution(row), nil
}

// toResolution maps a sqlc.Link to a redirect Resolution.
func toResolution(row sqlc.Link) *Resolution {
	return &Resolution{
		LinkID:  row.ID,
		DestURL: row.DestUrl,
	}
}
