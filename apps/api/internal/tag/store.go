package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Storer defines the interface for tag storage operations.
type Storer interface {
	Insert(ctx context.Context, t *Tag) (*Tag, error)
	FindByID(ctx context.Context, id, workspaceID string) (*Tag, error)
	List(ctx context.Context, workspaceID string) ([]Tag, error)
	Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Tag, error)
	SoftDelete(ctx context.Context, id, workspaceID string) error
	AddToLink(ctx context.Context, linkID, tagID string) error
	RemoveFromLink(ctx context.Context, linkID, tagID string) error
	ListByLink(ctx context.Context, linkID string) ([]Tag, error)
}

// Store handles tag database operations.
type Store struct {
	q *sqlc.Queries
}

// NewStore creates a new tag store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool)}
}

func (s *Store) Insert(ctx context.Context, t *Tag) (*Tag, error) {
	row, err := s.q.InsertTag(ctx, sqlc.InsertTagParams{
		ID:          t.ID,
		WorkspaceID: t.WorkspaceID,
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

func (s *Store) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Tag, error) {
	row, err := s.q.UpdateTag(ctx, sqlc.UpdateTagParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        input.Name,
		Color:       input.Color,
	})
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

func (s *Store) AddToLink(ctx context.Context, linkID, tagID string) error {
	if err := s.q.AddTagToLink(ctx, sqlc.AddTagToLinkParams{
		LinkID: linkID,
		TagID:  tagID,
	}); err != nil {
		return fmt.Errorf("adding tag %s to link %s: %w", tagID, linkID, err)
	}
	return nil
}

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

// toTag maps a sqlc.Tag to a domain Tag.
func toTag(row sqlc.Tag) *Tag {
	return &Tag{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Name:        row.Name,
		Color:       row.Color,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
