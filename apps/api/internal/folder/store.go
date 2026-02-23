package folder

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

// Store handles database operations for the folder package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new folder store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// Insert creates a new folder record.
func (s *Store) Insert(ctx context.Context, f *Folder) (*Folder, error) {
	row, err := s.q.InsertFolder(ctx, sqlc.InsertFolderParams{
		ID:          f.ID,
		WorkspaceID: f.WorkspaceID,
		CreatedBy:   ptr.ToNonZero(f.CreatedBy),
		Name:        f.Name,
		Color:       f.Color,
		Position:    ptr.ToInt32(f.Position),
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrNameTaken
		}
		return nil, fmt.Errorf("inserting folder: %w", err)
	}
	return toFolder(row), nil
}

// FindByID retrieves a single folder by ID.
func (s *Store) FindByID(ctx context.Context, id, workspaceID string) (*Folder, error) {
	row, err := s.q.FindFolderByID(ctx, sqlc.FindFolderByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying folder %s: %w", id, err)
	}
	return toFolder(row), nil
}

// List retrieves all active folders for a workspace.
func (s *Store) List(ctx context.Context, workspaceID string) ([]Folder, error) {
	rows, err := s.q.ListFoldersByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}

	folders := make([]Folder, len(rows))
	for i, row := range rows {
		folders[i] = *toFolder(row)
	}
	return folders, nil
}

// Update partially updates a folder.
func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Folder, error) {
	var u db.Update

	if input.Name.Set {
		u.Set("name", input.Name.Value)
	}
	if input.Color.Set {
		u.Set("color", input.Color.Value)
	}
	if input.Position.Set {
		u.Set("position", input.Position.Value)
	}

	if u.IsEmpty() {
		return s.FindByID(ctx, id, workspaceID)
	}

	sql, args := u.Build("folders", "id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Folder])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrNameTaken
		}
		return nil, fmt.Errorf("updating folder %s: %w", id, err)
	}
	return toFolder(row), nil
}

// SoftDelete marks a folder as deleted.
func (s *Store) SoftDelete(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteFolder(ctx, sqlc.SoftDeleteFolderParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting folder %s: %w", id, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func toFolder(row sqlc.Folder) *Folder {
	return &Folder{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		CreatedBy:   ptr.From(row.CreatedBy),
		Name:        row.Name,
		Color:       row.Color,
		Position:    int(row.Position),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
