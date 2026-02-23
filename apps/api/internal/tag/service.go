package tag

import (
	"context"

	"github.com/rs/xid"
)

const defaultColor = "#6366f1"

// Service implements tag business logic.
type Service struct {
	store *Store
}

// NewService creates a new tag service.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// Create persists a new tag.
func (s *Service) Create(ctx context.Context, workspaceID, userID string, input CreateInput) (*Tag, error) {
	color := input.Color
	if color == "" {
		color = defaultColor
	}

	t := &Tag{
		ID:          "tag_" + xid.New().String(),
		WorkspaceID: workspaceID,
		CreatedBy:   userID,
		Name:        input.Name,
		Color:       color,
	}

	return s.store.Insert(ctx, t)
}

// Get retrieves a tag by ID within a workspace.
func (s *Service) Get(ctx context.Context, id, workspaceID string) (*Tag, error) {
	return s.store.FindByID(ctx, id, workspaceID)
}

// List returns all tags for a workspace.
func (s *Service) List(ctx context.Context, workspaceID string) ([]Tag, error) {
	return s.store.List(ctx, workspaceID)
}

// Update partially updates a tag.
func (s *Service) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Tag, error) {
	return s.store.Update(ctx, id, workspaceID, input)
}

// Delete soft-deletes a tag.
func (s *Service) Delete(ctx context.Context, id, workspaceID string) error {
	return s.store.SoftDelete(ctx, id, workspaceID)
}

// AddToLink associates a tag with a link.
func (s *Service) AddToLink(ctx context.Context, linkID, tagID string) error {
	return s.store.AddToLink(ctx, linkID, tagID)
}

// RemoveFromLink removes a tag from a link.
func (s *Service) RemoveFromLink(ctx context.Context, linkID, tagID string) error {
	return s.store.RemoveFromLink(ctx, linkID, tagID)
}

// ListByLink returns all tags for a link.
func (s *Service) ListByLink(ctx context.Context, linkID string) ([]Tag, error) {
	return s.store.ListByLink(ctx, linkID)
}
