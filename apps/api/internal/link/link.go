// Package link handles URL shortening, including CRUD operations
// and short code generation.
package link

import (
	"crypto/rand"
	"errors"
	"time"
)

// Domain type.

// Link represents a shortened URL belonging to a workspace.
type Link struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	DestURL     string     `json:"dest_url"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	ClickCount  int64      `json:"click_count"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastClicked *time.Time `json:"last_clicked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Input types.

// CreateInput is the input for creating a new link.
type CreateInput struct {
	WorkspaceID string     `json:"workspace_id" validate:"required"`
	DestURL     string     `json:"dest_url"     validate:"required,url,max=2048"`
	ShortCode   string     `json:"short_code"   validate:"omitempty,min=1,max=50"`
	Title       string     `json:"title"        validate:"omitempty,max=200"`
	Description string     `json:"description"  validate:"omitempty,max=500"`
	ExpiresAt   *time.Time `json:"expires_at"   validate:"omitempty"`
}

// UpdateInput is the input for updating an existing link.
// Use patch.Bind[UpdateInput] to parse and track field presence.
type UpdateInput struct {
	DestURL     *string    `json:"dest_url"     validate:"omitempty,url,max=2048"`
	Title       *string    `json:"title"        validate:"omitempty,max=200"`
	Description *string    `json:"description"  validate:"omitempty,max=500"`
	IsActive    *bool      `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// Sentinel errors.

var (
	ErrNotFound       = errors.New("link not found")
	ErrShortCodeTaken = errors.New("short code already in use")
)

// Short code generation.

const (
	alphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCodeLen = 7
	maxRetries   = 3
	bitMask      = 63 // 6-bit mask (0-63); rejects 62,63 → ~3% rejection rate
)

// generateShortCode creates a cryptographically random short code.
// Uses a single rand.Read with bitmask rejection sampling for uniform
// distribution across 62 characters with minimal syscalls.
func generateShortCode() (string, error) {
	code := make([]byte, shortCodeLen)
	raw := make([]byte, shortCodeLen*2) // 2x buffer handles the ~3% rejection rate

	filled := 0
	for filled < shortCodeLen {
		if _, err := rand.Read(raw); err != nil {
			return "", err
		}

		for _, b := range raw {
			if idx := b & bitMask; idx < byte(len(alphabet)) {
				code[filled] = alphabet[idx]
				filled++
				if filled == shortCodeLen {
					break
				}
			}
		}
	}

	return string(code), nil
}
