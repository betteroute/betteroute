// Package tag handles link categorization via tags.
package tag

import (
	"errors"
	"time"

	"github.com/execrc/betteroute/internal/opt"
)

// Tag represents a workspace tag for categorizing links.
type Tag struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedBy   string    `json:"created_by,omitempty"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateInput is the input for creating a tag.
// WorkspaceID is securely injected by middleware.
type CreateInput struct {
	Name  string `json:"name"         validate:"required,min=1,max=50"`
	Color string `json:"color"        validate:"omitempty,hexcolor,len=7"`
}

// UpdateInput is the input for partially updating a tag.
type UpdateInput struct {
	Name  opt.Field[string] `json:"name"  validate:"omitempty,min=1,max=50" swaggertype:"string"`
	Color opt.Field[string] `json:"color" validate:"omitempty,hexcolor,len=7" swaggertype:"string"`
}

// AddToLinkInput is the input for associating a tag with a link.
type AddToLinkInput struct {
	TagID string `json:"tag_id" validate:"required"`
}

var (
	ErrNotFound  = errors.New("tag not found")
	ErrNameTaken = errors.New("tag name already exists")
)
