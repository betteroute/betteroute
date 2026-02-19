package folder

import (
	"context"

	"github.com/rs/xid"
)

const defaultColor = "#6366f1"

// Service handles folder business logic.
type Service struct {
	store *Store
}

// NewService creates a new folder service.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// Create persists a new folder.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Folder, error) {
	color := input.Color
	if color == "" {
		color = defaultColor
	}

	f := &Folder{
		ID:          "fld_" + xid.New().String(),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       color,
	}

	return s.store.Insert(ctx, f)
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
	return s.store.SoftDelete(ctx, id, workspaceID)
}
