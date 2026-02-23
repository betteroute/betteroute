// Package folder handles link organization via folders.
package folder

import (
	"errors"
	"time"

	"github.com/execrc/betteroute/internal/opt"
)

// Folder represents a workspace folder for organizing links.
type Folder struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedBy   string    `json:"created_by,omitempty"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateInput is the input for creating a folder.
// WorkspaceID is omitted because it is securely inferred from the route middleware.
type CreateInput struct {
	Name  string `json:"name"         validate:"required,min=1,max=100"`
	Color string `json:"color"        validate:"omitempty,hexcolor,len=7"`
}

// UpdateInput is the input for partially updating a folder.
type UpdateInput struct {
	Name     opt.Field[string] `json:"name"     validate:"omitempty,min=1,max=100" swaggertype:"string"`
	Color    opt.Field[string] `json:"color"    validate:"omitempty,hexcolor,len=7" swaggertype:"string"`
	Position opt.Field[*int32] `json:"position" swaggertype:"integer"`
}

var (
	ErrNotFound  = errors.New("folder not found")
	ErrNameTaken = errors.New("folder name already exists")
)
