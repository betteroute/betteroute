package redirect

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Service resolves short codes to redirect decisions.
type Service struct {
	q               *sqlc.Queries
	platformDomains map[string]bool
}

// NewService creates a new redirect service.
// platformDomains is the list of hostnames owned by the platform (e.g. "br.link").
func NewService(pool *pgxpool.Pool, platformDomains []string) *Service {
	pd := make(map[string]bool, len(platformDomains))
	for _, d := range platformDomains {
		pd[d] = true
	}
	return &Service{q: sqlc.New(pool), platformDomains: pd}
}

// Resolve looks up a short code, validates the link, and returns a Resolution.
//
// Domain-aware: custom domain requests scope the lookup to links assigned to
// that domain; platform domain requests match links with no domain_id.
//
// Happy path (1-2 DB round-trips): ResolveDomain (custom only) + ResolveLink.
// Fallback path (+1): when the UPDATE touches 0 rows, a slim SELECT diagnoses why.
func (s *Service) Resolve(ctx context.Context, code, hostname string) (*Resolution, error) {
	domainNS, err := s.resolveDomain(ctx, hostname)
	if err != nil {
		return nil, err
	}

	// Hot path: atomic UPDATE increments click_count and returns only the
	// columns needed for redirect in a single round-trip.
	row, err := s.q.ResolveLink(ctx, sqlc.ResolveLinkParams{
		ShortCode: code,
		DomainNs:  domainNS,
	})
	if err == nil {
		return toResolution(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err // unexpected DB error
	}

	// Cold path: the UPDATE matched nothing — the link is either missing,
	// inactive, expired, not yet started, or click-limited. A slim SELECT
	// diagnoses the reason so we can serve expiration_url fallbacks.
	return s.diagnoseFallback(ctx, code, domainNS)
}

// resolveDomain returns the domain namespace for the ResolveLink query.
// Platform domains → "" (matches links with NULL domain_id).
// Custom domains → domain ID from the domains table.
func (s *Service) resolveDomain(ctx context.Context, hostname string) (string, error) {
	if s.platformDomains[hostname] {
		return "", nil
	}

	dom, err := s.q.ResolveDomain(ctx, hostname)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound // unknown hostname
	}
	if err != nil {
		return "", err
	}
	return dom.ID, nil
}

// diagnoseFallback determines why ResolveLink returned no rows and returns
// the appropriate error or an expiration_url redirect.
func (s *Service) diagnoseFallback(ctx context.Context, code, domainNS string) (*Resolution, error) {
	fb, err := s.q.FindRedirectFallback(ctx, sqlc.FindRedirectFallbackParams{
		ShortCode: code,
		DomainNs:  domainNS,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound // short code doesn't exist at all
		}
		return nil, err
	}

	// Walk the gate checks in the same order as the UPDATE's WHERE clause
	// to return the most specific reason for rejection.

	if !fb.IsActive {
		return nil, ErrInactive
	}

	now := time.Now()

	if fb.StartsAt != nil && now.Before(*fb.StartsAt) {
		return nil, ErrNotStarted
	}

	if fb.ExpiresAt != nil && now.After(*fb.ExpiresAt) {
		return resolveExpiration(fb.ExpirationUrl, ErrExpired)
	}

	if fb.MaxClicks != nil && fb.ClickCount >= int64(*fb.MaxClicks) {
		return resolveExpiration(fb.ExpirationUrl, ErrClickLimitReached)
	}

	// Shouldn't reach here — gate checks mirror the UPDATE WHERE clause.
	return nil, ErrNotFound
}

// resolveExpiration redirects to expiration_url if set, otherwise returns the
// given sentinel error. Used for both expired and click-limited links.
func resolveExpiration(expirationURL *string, sentinel error) (*Resolution, error) {
	if expirationURL != nil && *expirationURL != "" {
		return &Resolution{DestURL: *expirationURL}, nil
	}
	return nil, sentinel
}

// EnrichDeepLinks populates the Resolution's deep link fields from the
// deep_links table. Called by the handler only when the device is mobile.
func (s *Service) EnrichDeepLinks(ctx context.Context, res *Resolution) {
	row, err := s.q.ResolveDeepLink(ctx, res.LinkID)
	if err != nil {
		return // no deep link data — not an error
	}

	res.IOSDeepLink = ptr.Val(row.IosDeepLink)
	res.AndroidDeepLink = ptr.Val(row.AndroidDeepLink)
	res.IOSFallbackURL = ptr.Val(row.IosFallbackUrl)
	res.AndroidFallbackURL = ptr.Val(row.AndroidFallbackUrl)
	res.AndroidPackage = ptr.Val(row.AndroidPackage)
}

// toResolution maps a sqlc.ResolveLinkRow to a redirect Resolution.
func toResolution(row sqlc.ResolveLinkRow) *Resolution {
	dest := appendUTM(row.DestUrl, row)

	return &Resolution{
		LinkID:        row.ID,
		WorkspaceID:   row.WorkspaceID,
		DestURL:       dest,
		OGTitle:       ptr.Val(row.OgTitle),
		OGDescription: ptr.Val(row.OgDescription),
		OGImage:       ptr.Val(row.OgImage),
	}
}

// appendUTM appends stored UTM params to the destination URL query string.
func appendUTM(dest string, row sqlc.ResolveLinkRow) string {
	params := []struct {
		key string
		val *string
	}{
		{"utm_source", row.UtmSource},
		{"utm_medium", row.UtmMedium},
		{"utm_campaign", row.UtmCampaign},
		{"utm_term", row.UtmTerm},
		{"utm_content", row.UtmContent},
	}

	// Fast exit: skip URL parsing when no UTM params are set (common case).
	hasAny := false
	for _, p := range params {
		if p.val != nil && *p.val != "" {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return dest
	}

	u, err := url.Parse(dest)
	if err != nil {
		return dest // malformed dest — return as-is, don't block the redirect
	}

	q := u.Query()
	for _, p := range params {
		if p.val != nil && *p.val != "" {
			q.Set(p.key, *p.val)
		}
	}
	u.RawQuery = q.Encode()

	return u.String()
}
