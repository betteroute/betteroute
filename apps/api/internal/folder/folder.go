// Package folder handles link organization via folders.
package folder

import (
	"errors"
	"time"
)

// Domain type.

// Folder represents a workspace folder for organizing links.
type Folder struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Input types.

// CreateInput is the input for creating a folder.
type CreateInput struct {
	WorkspaceID string `json:"workspace_id" validate:"required"`
	Name        string `json:"name"         validate:"required,min=1,max=100"`
	Color       string `json:"color"        validate:"omitempty,hexcolor,len=7"`
}

// UpdateInput is the input for updating a folder.
type UpdateInput struct {
	Name     *string `json:"name"     validate:"omitempty,min=1,max=100"`
	Color    *string `json:"color"    validate:"omitempty,hexcolor,len=7"`
	Position *int32  `json:"position"`
}

// NullableFields tracks which nullable fields should be explicitly set (vs ignored).
type NullableFields struct {
	Position bool
}

// Sentinel errors.

var (
	ErrNotFound  = errors.New("folder not found")
	ErrNameTaken = errors.New("folder name already exists")
)
