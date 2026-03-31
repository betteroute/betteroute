package folder

import (
	"context"
	"log/slog"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/usage"
)

const defaultColor = "#6366f1"

// Service implements folder business logic.
type Service struct {
	store *Store
	meter *usage.Meter
}

// NewService creates a new folder service.
func NewService(store *Store, meter *usage.Meter) *Service {
	return &Service{store: store, meter: meter}
}

// Create persists a new folder.
func (s *Service) Create(ctx context.Context, workspaceID, userID string, input CreateInput) (*Folder, error) {
	color := input.Color
	if color == "" {
		color = defaultColor
	}

	f := &Folder{
		ID:          "fld_" + xid.New().String(),
		WorkspaceID: workspaceID,
		CreatedBy:   userID,
		Name:        input.Name,
		Color:       color,
	}

	created, err := s.store.Insert(ctx, f)
	if err != nil {
		return nil, err
	}

	if err := s.meter.Adjust(ctx, workspaceID, usage.Folders, 1); err != nil {
		slog.WarnContext(ctx, "adjusting folder usage", "error", err, "workspace_id", workspaceID)
	}

	return created, nil
}

// Get retrieves a folder by ID within a workspace.
func (s *Service) Get(ctx context.Context, id, workspaceID string) (*Folder, error) {
	return s.store.FindByID(ctx, id, workspaceID)
}

// List returns all folders for a workspace.
func (s *Service) List(ctx context.Context, workspaceID string) ([]Folder, error) {
	return s.store.List(ctx, workspaceID)
}

// Update partially updates a folder.
func (s *Service) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Folder, error) {
	return s.store.Update(ctx, id, workspaceID, input)
}

// Delete soft-deletes a folder.
func (s *Service) Delete(ctx context.Context, id, workspaceID string) error {
	if err := s.store.SoftDelete(ctx, id, workspaceID); err != nil {
		return err
	}

	if err := s.meter.Adjust(ctx, workspaceID, usage.Folders, -1); err != nil {
		slog.WarnContext(ctx, "adjusting folder usage", "error", err, "workspace_id", workspaceID)
	}

	return nil
}
