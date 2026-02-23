package tag

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

// Store handles database operations for the tag package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new tag store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// Insert creates a new tag record.
func (s *Store) Insert(ctx context.Context, t *Tag) (*Tag, error) {
	row, err := s.q.InsertTag(ctx, sqlc.InsertTagParams{
		ID:          t.ID,
		WorkspaceID: t.WorkspaceID,
		CreatedBy:   ptr.ToNonZero(t.CreatedBy),
		Name:        t.Name,
		Color:       t.Color,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrNameTaken
		}
		return nil, fmt.Errorf("inserting tag: %w", err)
	}
	return toTag(row), nil
}

// FindByID retrieves a single tag by ID.
func (s *Store) FindByID(ctx context.Context, id, workspaceID string) (*Tag, error) {
	row, err := s.q.FindTagByID(ctx, sqlc.FindTagByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying tag %s: %w", id, err)
	}
	return toTag(row), nil
}

// List retrieves all active tags for a workspace.
func (s *Store) List(ctx context.Context, workspaceID string) ([]Tag, error) {
	rows, err := s.q.ListTagsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	tags := make([]Tag, len(rows))
	for i, row := range rows {
		tags[i] = *toTag(row)
	}
	return tags, nil
}

// Update partially updates a tag.
func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Tag, error) {
	var u db.Update

	if input.Name.Set {
		u.Set("name", input.Name.Value)
	}
	if input.Color.Set {
		u.Set("color", input.Color.Value)
	}

	if u.IsEmpty() {
		return s.FindByID(ctx, id, workspaceID)
	}

	sql, args := u.Build("tags", "id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.Tag])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrNameTaken
		}
		return nil, fmt.Errorf("updating tag %s: %w", id, err)
	}
	return toTag(row), nil
}

// SoftDelete marks a tag as deleted.
func (s *Store) SoftDelete(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteTag(ctx, sqlc.SoftDeleteTagParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting tag %s: %w", id, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// AddToLink associates a tag with a link.
func (s *Store) AddToLink(ctx context.Context, linkID, tagID string) error {
	if err := s.q.AddTagToLink(ctx, sqlc.AddTagToLinkParams{
		LinkID: linkID,
		TagID:  tagID,
	}); err != nil {
		return fmt.Errorf("adding tag %s to link %s: %w", tagID, linkID, err)
	}
	return nil
}

// RemoveFromLink removes a tag association from a link.
func (s *Store) RemoveFromLink(ctx context.Context, linkID, tagID string) error {
	rows, err := s.q.RemoveTagFromLink(ctx, sqlc.RemoveTagFromLinkParams{
		LinkID: linkID,
		TagID:  tagID,
	})
	if err != nil {
		return fmt.Errorf("removing tag %s from link %s: %w", tagID, linkID, err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByLink retrieves all active tags associated with a specific link.
func (s *Store) ListByLink(ctx context.Context, linkID string) ([]Tag, error) {
	rows, err := s.q.ListTagsByLink(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("listing tags for link %s: %w", linkID, err)
	}

	tags := make([]Tag, len(rows))
	for i, row := range rows {
		tags[i] = *toTag(row)
	}
	return tags, nil
}

func toTag(row sqlc.Tag) *Tag {
	return &Tag{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		CreatedBy:   ptr.From(row.CreatedBy),
		Name:        row.Name,
		Color:       row.Color,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
