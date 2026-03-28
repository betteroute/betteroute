// Package redirect handles resolving short codes to their destination URLs.
// The service queries sqlc directly (no store layer) — this is the redirect hot
// path and the extra abstraction adds no value for a read-heavy, write-light flow.
package redirect

import "errors"

// Resolution is the terminal decision of a redirect request.
// The handler acts on the result; the service decides what it contains.
type Resolution struct {
	LinkID      string
	WorkspaceID string
	DestURL     string

	// OG metadata — served as HTML to social crawlers for rich previews.
	OGTitle       string
	OGDescription string
	OGImage       string

	// Deep link data — populated by EnrichDeepLinks when the handler
	// decides a deepview interstitial is needed (mobile + in-app browser).
	IOSDeepLink        string
	AndroidDeepLink    string
	IOSFallbackURL     string // App Store URL or dest_url
	AndroidFallbackURL string // Play Store URL or dest_url
	AndroidPackage     string // for building intent:// URLs
}

// HasOG returns true if any Open Graph metadata is set.
func (r *Resolution) HasOG() bool {
	return r.OGTitle != "" || r.OGDescription != "" || r.OGImage != ""
}

// HasDeepLinks returns true if any deep link URL is set.
func (r *Resolution) HasDeepLinks() bool {
	return r.IOSDeepLink != "" || r.AndroidDeepLink != ""
}

// Sentinel errors.
var (
	ErrNotFound          = errors.New("link not found")
	ErrInactive          = errors.New("link is inactive")
	ErrExpired           = errors.New("link has expired")
	ErrNotStarted        = errors.New("link has not started yet")
	ErrClickLimitReached = errors.New("link click limit reached")
)
