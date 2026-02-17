package redirect

import "errors"

// Resolution is the terminal decision of a redirect request.
// The handler acts on the result; the service decides what it contains.
type Resolution struct {
	LinkID  string
	DestURL string
}

// Resolver resolves a short code to a redirect decision.
type Resolver interface {
	Resolve(code string) (*Resolution, error)
}

// Sentinel errors.
var (
	ErrNotFound = errors.New("link not found")
	ErrInactive = errors.New("link is inactive")
	ErrExpired  = errors.New("link has expired")
)
