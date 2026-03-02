package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database operations for the domain package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new domain store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// Insert creates a new domain record.
func (s *Store) Insert(ctx context.Context, d *Domain) (*Domain, error) {
	row, err := s.q.InsertDomain(ctx, sqlc.InsertDomainParams{
		ID:                d.ID,
		WorkspaceID:       d.WorkspaceID,
		CreatedBy:         ptr.ToNonZero(d.CreatedBy),
		Hostname:          d.Hostname,
		VerificationToken: d.VerificationToken,
		FallbackUrl:       ptr.ToNonZero(d.FallbackURL),
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrHostnameTaken
		}
		return nil, fmt.Errorf("inserting domain: %w", err)
	}
	return toDomain(row), nil
}

// FindByID retrieves a single domain by ID.
func (s *Store) FindByID(ctx context.Context, id, workspaceID string) (*Domain, error) {
	row, err := s.q.FindDomainByID(ctx, sqlc.FindDomainByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying domain %s: %w", id, err)
	}
	return toDomain(row), nil
}

// FindByHostname retrieves a domain by hostname (redirect hot path).
func (s *Store) FindByHostname(ctx context.Context, hostname string) (*Domain, error) {
	row, err := s.q.FindDomainByHostname(ctx, hostname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying domain by hostname %s: %w", hostname, err)
	}
	return toDomain(row), nil
}

// List retrieves all active domains for a workspace.
func (s *Store) List(ctx context.Context, workspaceID string) ([]Domain, error) {
	rows, err := s.q.ListDomainsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing domains: %w", err)
	}

	domains := make([]Domain, len(rows))
	for i, row := range rows {
		domains[i] = *toDomain(row)
	}
	return domains, nil
}

// Update partially updates a domain.
func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Domain, error) {
	var u db.Update

	if input.FallbackURL.Set {
		u.Set("fallback_url", input.FallbackURL.Value)
	}

	if u.IsEmpty() {
		return s.FindByID(ctx, id, workspaceID)
	}

	sql, args := u.Build("domains", "id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID)
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("updating domain %s: %w", id, err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Domain])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating domain %s: %w", id, err)
	}
	return toDomain(row), nil
}

// UpdateStatus sets the domain status and verified_at timestamp.
func (s *Store) UpdateStatus(ctx context.Context, id, workspaceID, status string, verifiedAt *time.Time) (*Domain, error) {
	row, err := s.q.UpdateDomainStatus(ctx, sqlc.UpdateDomainStatusParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Status:      status,
		VerifiedAt:  verifiedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating domain status %s: %w", id, err)
	}
	return toDomain(row), nil
}

// SoftDelete marks a domain as deleted.
func (s *Store) SoftDelete(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteDomain(ctx, sqlc.SoftDeleteDomainParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting domain %s: %w", id, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func toDomain(row sqlc.Domain) *Domain {
	return &Domain{
		ID:                row.ID,
		WorkspaceID:       row.WorkspaceID,
		CreatedBy:         ptr.From(row.CreatedBy),
		Hostname:          row.Hostname,
		VerificationToken: row.VerificationToken,
		VerifiedAt:        row.VerifiedAt,
		FallbackURL:       ptr.From(row.FallbackUrl),
		Status:            row.Status,
		LastCheckedAt:     row.LastCheckedAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
