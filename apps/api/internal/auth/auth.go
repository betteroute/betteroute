// Package auth handles user authentication: magic link login, OAuth,
// session management, and profile updates.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/opt"
)

// User is an authenticated identity.
type User struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	Status          string     `json:"status"`
	OnboardedAt     *time.Time `json:"onboarded_at,omitempty"`
	Timezone        string     `json:"timezone"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Session represents an active user session.
// Token holds the plain opaque value sent to the client once on creation — never persisted.
// UserID, IPAddress, UserAgent are internal and excluded from JSON responses.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Token     string    `json:"-"` // set on creation, written to cookie, cleared after
	ExpiresAt time.Time `json:"-"` // set in cookie, not exposed in body
	IPAddress string    `json:"-"`
	UserAgent string    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

// Account links a user to an auth provider.
// Internal auth plumbing — not exposed in API responses.
type Account struct {
	ID                string
	UserID            string
	Provider          string // "email" | "google" | "github"
	ProviderAccountID string
}

// VerificationToken is a one-time token for magic link authentication.
// Internal auth plumbing — not exposed in API responses.
// PlainToken is set before Insert, used by the store to compute and persist the hash.
type VerificationToken struct {
	ID         string
	UserID     string
	Email      string
	PlainToken string // set by caller; store hashes before persisting
	Type       string // "magic_link"

	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// SessionMeta carries server-derived HTTP metadata for session creation.
// Extracted by the handler from the request — never user-supplied.
type SessionMeta struct {
	IPAddress string
	UserAgent string
}

// Config holds configuration for the auth service and handler.
type Config struct {
	WebURL string // frontend URL — used in email links and post-OAuth redirects
	APIURL string // this API's base URL — used for OAuth callback URLs

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
}

// MagicLinkInput requests a one-time login link.
type MagicLinkInput struct {
	Email string `json:"email" validate:"required,email,max=254"`
	Name  string `json:"name"  validate:"omitempty,max=100"` // optional for new users
}

// VerifyMagicLinkInput consumes the magic link token to create a session.
type VerifyMagicLinkInput struct {
	Token string `json:"token" validate:"required"`
}

// UpdateProfileInput is the payload for updating the authenticated user's profile.
type UpdateProfileInput struct {
	Name      opt.Field[string]  `json:"name"       validate:"omitempty,min=1,max=100" swaggertype:"string"`
	AvatarURL opt.Field[*string] `json:"avatar_url" validate:"omitempty,url,max=2048" swaggertype:"string"`
	Timezone  opt.Field[string]  `json:"timezone"   validate:"omitempty,max=100" swaggertype:"string"`
}

// Context bundles the authenticated user and their active session.
// Injected by the Auth middleware.
type Context struct {
	User    *User
	Session *Session
}

type contextKey struct{}

// NewContext returns a copy of parent with the auth Context attached.
func NewContext(parent context.Context, actx Context) context.Context {
	return context.WithValue(parent, contextKey{}, actx)
}

// FromContext extracts the authenticated context.
// Returns a zero Context outside authenticated routes.
func FromContext(ctx context.Context) Context {
	actx, _ := ctx.Value(contextKey{}).(Context)
	return actx
}

// Sentinel errors.

var (
	// ErrNotFound is returned when a requested user or resource does not exist.
	ErrNotFound          = errors.New("user not found")
	ErrEmailTaken        = errors.New("email already in use")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrSessionNotFound   = errors.New("session not found")
	ErrTokenInvalid      = errors.New("token is invalid or expired")
	ErrEmailUnverified   = errors.New("email not verified")
	ErrAccountSuspended  = errors.New("account suspended")
	ErrRateLimited       = errors.New("too many requests, try again later")
	ErrOAuthUnavailable  = errors.New("oauth provider not configured")
)

// ID generators — prefixed for readability in logs and debugging.

func newUserID() string    { return "usr_" + xid.New().String() }
func newSessionID() string { return "ses_" + xid.New().String() }
func newAccountID() string { return "acc_" + xid.New().String() }
func newTokenID() string   { return "vtk_" + xid.New().String() }

// generateToken creates a cryptographically secure opaque token.
// Returns the plain token (sent to client / email) and its SHA-256 hash (stored in DB).
func generateToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b) // 64-char hex string
	hash = hashToken(plain)
	return
}

// hashToken computes the SHA-256 hash of a plain token for DB lookup.
func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
