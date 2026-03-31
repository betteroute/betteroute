package tag

import (
	"context"
	"log/slog"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/usage"
)

const defaultColor = "#6366f1"

// Service implements tag business logic.
type Service struct {
	store *Store
	meter *usage.Meter
}

// NewService creates a new tag service.
func NewService(store *Store, meter *usage.Meter) *Service {
	return &Service{store: store, meter: meter}
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

	created, err := s.store.Insert(ctx, t)
	if err != nil {
		return nil, err
	}

	if err := s.meter.Adjust(ctx, workspaceID, usage.Tags, 1); err != nil {
		slog.WarnContext(ctx, "adjusting tag usage", "error", err, "workspace_id", workspaceID)
	}

	return created, nil
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
	if err := s.store.SoftDelete(ctx, id, workspaceID); err != nil {
		return err
	}

	if err := s.meter.Adjust(ctx, workspaceID, usage.Tags, -1); err != nil {
		slog.WarnContext(ctx, "adjusting tag usage", "error", err, "workspace_id", workspaceID)
	}

	return nil
}

// AddToLink associates a tag with a link.
func (s *Service) AddToLink(ctx context.Context, linkID, tagID, workspaceID string) error {
	return s.store.AddToLink(ctx, linkID, tagID, workspaceID)
}

// RemoveFromLink removes a tag from a link.
func (s *Service) RemoveFromLink(ctx context.Context, linkID, tagID, workspaceID string) error {
	return s.store.RemoveFromLink(ctx, linkID, tagID, workspaceID)
}

// ListByLink returns all tags for a link.
func (s *Service) ListByLink(ctx context.Context, linkID, workspaceID string) ([]Tag, error) {
	return s.store.ListByLink(ctx, linkID, workspaceID)
}
