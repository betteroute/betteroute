package wellknown

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database queries for the wellknown package.
type Store struct {
	q *sqlc.Queries
}

// NewStore creates a new wellknown store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

// FindWorkspaceApps resolves hostname → workspace_id → workspace apps filtered by platform.
// Returns nil for unknown domains or domains with no matching apps.
func (s *Store) FindWorkspaceApps(ctx context.Context, hostname, platform string) ([]WorkspaceApp, error) {
	domain, err := s.q.ResolveDomain(ctx, hostname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolving domain %s: %w", hostname, err)
	}

	rows, err := s.q.ListWorkspaceAppsByPlatform(ctx, sqlc.ListWorkspaceAppsByPlatformParams{
		WorkspaceID: domain.WorkspaceID,
		Platform:    platform,
	})
	if err != nil {
		return nil, fmt.Errorf("listing %s apps for domain %s: %w", platform, hostname, err)
	}

	apps := make([]WorkspaceApp, len(rows))
	for i, row := range rows {
		apps[i] = WorkspaceApp{
			TeamID:             row.TeamID,
			BundleID:           row.BundleID,
			PackageName:        row.PackageName,
			SHA256Fingerprints: row.Sha256Fingerprints,
		}
	}
	return apps, nil
}
