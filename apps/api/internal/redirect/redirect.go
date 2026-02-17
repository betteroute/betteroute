package redirect

import "errors"

// Resolution is the terminal decision of a redirect request.
// The handler acts on the result; the service decides what it contains.
type Resolution struct {
	LinkID  string
	DestURL string

	// OG metadata — populated when the link has custom social previews.
	// The handler uses these to serve HTML to social crawlers.
	OGTitle       string
	OGDescription string
	OGImage       string
}

// HasOG returns true if any Open Graph metadata is set.
func (r *Resolution) HasOG() bool {
	return r.OGTitle != "" || r.OGDescription != "" || r.OGImage != ""
}

// Resolver resolves a short code to a redirect decision.
type Resolver interface {
	Resolve(code string) (*Resolution, error)
}

// Sentinel errors.
var (
	ErrNotFound          = errors.New("link not found")
	ErrInactive          = errors.New("link is inactive")
	ErrExpired           = errors.New("link has expired")
	ErrNotStarted        = errors.New("link has not started yet")
	ErrClickLimitReached = errors.New("link click limit reached")
)
