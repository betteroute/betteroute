package apikey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database operations for the apikey package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new API key store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// InsertParams holds the fields needed to persist a new API key.
type InsertParams struct {
	ID          string
	WorkspaceID string
	CreatedBy   *string
	Name        string
	KeyHash     string
	KeyPrefix   string
	Permission  Permission
	Scopes      []string
	ExpiresAt   *time.Time
}

// Insert creates a new API key record.
func (s *Store) Insert(ctx context.Context, p InsertParams) (*APIKey, error) {
	row, err := s.q.InsertAPIKey(ctx, sqlc.InsertAPIKeyParams{
		ID:          p.ID,
		WorkspaceID: p.WorkspaceID,
		CreatedBy:   p.CreatedBy,
		Name:        p.Name,
		KeyHash:     p.KeyHash,
		KeyPrefix:   p.KeyPrefix,
		Permission:  string(p.Permission),
		Scopes:      p.Scopes,
		ExpiresAt:   p.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("inserting api key: %w", err)
	}
	return toAPIKey(row), nil
}

// FindByHash resolves a key hash to its domain object (auth hot path).
func (s *Store) FindByHash(ctx context.Context, hash string) (*APIKey, error) {
	row, err := s.q.FindAPIKeyByHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying api key by hash: %w", err)
	}
	return toAPIKey(row), nil
}

// FindByID retrieves an API key by its ID.
func (s *Store) FindByID(ctx context.Context, id, workspaceID string) (*APIKey, error) {
	row, err := s.q.FindAPIKeyByID(ctx, sqlc.FindAPIKeyByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying api key %s: %w", id, err)
	}
	return toAPIKey(row), nil
}

// List retrieves all active API keys for a workspace.
func (s *Store) List(ctx context.Context, workspaceID string) ([]APIKey, error) {
	rows, err := s.q.ListAPIKeysByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}

	keys := make([]APIKey, len(rows))
	for i, row := range rows {
		keys[i] = *toAPIKey(row)
	}
	return keys, nil
}

// SoftDelete marks an API key as deleted.
func (s *Store) SoftDelete(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteAPIKey(ctx, sqlc.SoftDeleteAPIKeyParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting api key %s: %w", id, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateLastUsed updates the last_used_at timestamp.
func (s *Store) UpdateLastUsed(ctx context.Context, id string) error {
	if err := s.q.UpdateAPIKeyLastUsed(ctx, id); err != nil {
		return fmt.Errorf("updating last used for api key %s: %w", id, err)
	}
	return nil
}

func toAPIKey(row sqlc.ApiKey) *APIKey {
	return &APIKey{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		CreatedBy:   ptr.From(row.CreatedBy),
		Name:        row.Name,
		KeyPrefix:   row.KeyPrefix,
		Permission:  Permission(row.Permission),
		Scopes:      stringsToScopes(row.Scopes),
		ExpiresAt:   row.ExpiresAt,
		LastUsedAt:  row.LastUsedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
