// Package link handles URL shortening, including CRUD operations
// and short code generation.
package link

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/execrc/betteroute/internal/opt"
)

// Link represents a shortened URL belonging to a workspace.
type Link struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	CreatedBy   string `json:"created_by,omitempty"`
	FolderID    string `json:"folder_id,omitempty"`
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	DestURL     string `json:"dest_url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Status & Scheduling
	IsActive      bool       `json:"is_active"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ExpirationURL string     `json:"expiration_url,omitempty"`

	// Click limits
	MaxClicks *int32 `json:"max_clicks,omitempty"`

	// UTM parameters
	UTMSource   string `json:"utm_source,omitempty"`
	UTMMedium   string `json:"utm_medium,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
	UTMTerm     string `json:"utm_term,omitempty"`
	UTMContent  string `json:"utm_content,omitempty"`

	// OG metadata overrides
	OGTitle       string `json:"og_title,omitempty"`
	OGDescription string `json:"og_description,omitempty"`
	OGImage       string `json:"og_image,omitempty"`

	// Analytics (denormalized)
	ClickCount       int64      `json:"click_count"`
	UniqueClickCount int64      `json:"unique_click_count"`
	LastClicked      *time.Time `json:"last_clicked_at,omitempty"`

	// Internal
	Notes      string `json:"notes,omitempty"`
	CreatedVia string `json:"created_via"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateInput is the input for creating a new link.
// WorkspaceID is securely injected by middleware.
type CreateInput struct {
	FolderID       string `json:"folder_id"        validate:"omitempty"`
	WorkspaceAppID string `json:"workspace_app_id" validate:"omitempty"`
	DestURL        string `json:"dest_url"         validate:"required,url,max=2048"`
	ShortCode      string `json:"short_code"       validate:"omitempty,min=1,max=50,shortcode"`
	Title          string `json:"title"        validate:"omitempty,max=200"`
	Description    string `json:"description"  validate:"omitempty,max=500"`

	// Scheduling
	StartsAt      *time.Time `json:"starts_at"       validate:"omitempty"`
	ExpiresAt     *time.Time `json:"expires_at"      validate:"omitempty"`
	ExpirationURL string     `json:"expiration_url"  validate:"omitempty,url,max=2048"`

	// Click limits
	MaxClicks *int32 `json:"max_clicks" validate:"omitempty,gt=0"`

	// UTM parameters
	UTMSource   string `json:"utm_source"   validate:"omitempty,max=200"`
	UTMMedium   string `json:"utm_medium"   validate:"omitempty,max=200"`
	UTMCampaign string `json:"utm_campaign" validate:"omitempty,max=200"`
	UTMTerm     string `json:"utm_term"     validate:"omitempty,max=200"`
	UTMContent  string `json:"utm_content"  validate:"omitempty,max=200"`

	// OG metadata overrides
	OGTitle       string `json:"og_title"       validate:"omitempty,max=200"`
	OGDescription string `json:"og_description" validate:"omitempty,max=500"`
	OGImage       string `json:"og_image"       validate:"omitempty,url,max=2048"`

	// Internal
	Notes string `json:"notes" validate:"omitempty,max=5000"`
}

// UpdateInput is the input for partially updating a link.
// Fields use opt.Field to track presence for PATCH semantics.
type UpdateInput struct {
	FolderID       opt.Field[*string] `json:"folder_id"        swaggertype:"string"`
	WorkspaceAppID opt.Field[*string] `json:"workspace_app_id" swaggertype:"string"`
	DestURL        opt.Field[*string] `json:"dest_url"         validate:"omitempty,url,max=2048" swaggertype:"string"`
	Title          opt.Field[*string] `json:"title"            validate:"omitempty,max=200" swaggertype:"string"`
	Description    opt.Field[*string] `json:"description"  validate:"omitempty,max=500" swaggertype:"string"`
	IsActive       opt.Field[*bool]   `json:"is_active" swaggertype:"boolean"`

	// Scheduling
	StartsAt      opt.Field[*time.Time] `json:"starts_at" swaggertype:"string"`
	ExpiresAt     opt.Field[*time.Time] `json:"expires_at" swaggertype:"string"`
	ExpirationURL opt.Field[*string]    `json:"expiration_url" validate:"omitempty,url,max=2048" swaggertype:"string"`

	// Click limits
	MaxClicks opt.Field[*int32] `json:"max_clicks" validate:"omitempty,gt=0" swaggertype:"integer"`

	// UTM parameters
	UTMSource   opt.Field[*string] `json:"utm_source"   validate:"omitempty,max=200" swaggertype:"string"`
	UTMMedium   opt.Field[*string] `json:"utm_medium"   validate:"omitempty,max=200" swaggertype:"string"`
	UTMCampaign opt.Field[*string] `json:"utm_campaign" validate:"omitempty,max=200" swaggertype:"string"`
	UTMTerm     opt.Field[*string] `json:"utm_term"     validate:"omitempty,max=200" swaggertype:"string"`
	UTMContent  opt.Field[*string] `json:"utm_content"  validate:"omitempty,max=200" swaggertype:"string"`

	// OG metadata overrides
	OGTitle       opt.Field[*string] `json:"og_title"       validate:"omitempty,max=200" swaggertype:"string"`
	OGDescription opt.Field[*string] `json:"og_description" validate:"omitempty,max=500" swaggertype:"string"`
	OGImage       opt.Field[*string] `json:"og_image"       validate:"omitempty,url,max=2048" swaggertype:"string"`

	// Internal
	Notes opt.Field[*string] `json:"notes" validate:"omitempty,max=5000" swaggertype:"string"`
}

var (
	ErrNotFound       = errors.New("link not found")
	ErrShortCodeTaken = errors.New("short code already in use")
)

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
