package redirect

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"

	"github.com/gofiber/fiber/v3"
	useragent "github.com/medama-io/go-useragent"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/usage"
)

//go:embed templates/og.html
var ogHTML string

//go:embed templates/deepview.html
var deepviewHTML string

// ogTmpl is the parsed OG template, initialized once at package load.
var ogTmpl = template.Must(template.New("og").Parse(ogHTML))

// deepviewTmpl is the parsed deepview interstitial template.
var deepviewTmpl = template.Must(template.New("deepview").Parse(deepviewHTML))

// Handler handles short code redirect requests.
type Handler struct {
	svc   *Service
	ua    *useragent.Parser
	meter *usage.Meter
}

// NewHandler creates a new redirect handler.
func NewHandler(svc *Service, meter *usage.Meter) *Handler {
	return &Handler{
		svc:   svc,
		ua:    useragent.NewParser(),
		meter: meter,
	}
}

// Register mounts the catch-all redirect route.
// Must be registered LAST to avoid catching /api/v1 and /health.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/:code", h.Redirect)
}

// @Summary     Redirect short link
// @Description Resolves a short code and redirects to the destination URL.
// @Tags        redirect
// @Param       code path string true "Short code"
// @Success     302  "Redirect to destination"
// @Failure     404  {object} errs.Error "Link not found, inactive, or expired"
// @Failure     500  {object} errs.Error
// @Router      /{code} [get]
func (h *Handler) Redirect(c fiber.Ctx) error {
	code := c.Params("code")

	res, err := h.svc.Resolve(c.Context(), code, c.Hostname())
	if err != nil {
		return mapError(err)
	}

	// Track click against workspace quota (async, non-blocking).
	h.meter.Emit(res.WorkspaceID, usage.Clicks, 1)

	// Serve OG HTML to social crawlers for rich preview cards.
	// We serve this even if the user didn't set custom OG data, because the
	// template has sensible fallbacks (like the destination URL).
	if h.ua.Parse(c.Get("User-Agent")).IsBot() {
		return h.serveOG(c, res)
	}

	// Deep linking: enrich + serve deepview for mobile in-app browsers.
	device := DetectDevice(c.Get("User-Agent"))
	if device.IsMobile() {
		h.svc.EnrichDeepLinks(c.Context(), res)

		if res.HasDeepLinks() && device.IsInApp() {
			return h.serveDeepview(c, res, device)
		}
	}

	c.Set("Cache-Control", "private, no-store")
	return c.Redirect().Status(fiber.StatusFound).To(res.DestURL)
}

// serveOG renders the OG template and returns it as HTML.
func (h *Handler) serveOG(c fiber.Ctx, res *Resolution) error {
	var buf bytes.Buffer
	if err := ogTmpl.Execute(&buf, res); err != nil {
		return errs.Internal("").WithCause(err)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "public, max-age=300")
	return c.Send(buf.Bytes())
}

// deepviewConfig is the JSON config passed to the deepview template's JS.
type deepviewConfig struct {
	OS              string `json:"os"`
	DestURL         string `json:"destURL"`
	IOSDeepLink     string `json:"iosDeepLink,omitempty"`
	AndroidDeepLink string `json:"androidDeepLink,omitempty"`
	IOSFallback     string `json:"iosFallback,omitempty"`
	AndroidFallback string `json:"androidFallback,omitempty"`
	AndroidPackage  string `json:"androidPackage,omitempty"`
}

// deepviewData is the template data for the deepview interstitial.
type deepviewData struct {
	Title       string
	Description string
	OGImage     string
	DestURL     string
	ConfigJSON  string
}

// serveDeepview renders the deepview interstitial page for in-app browsers.
func (h *Handler) serveDeepview(c fiber.Ctx, res *Resolution, device DeviceInfo) error {
	cfg := deepviewConfig{
		OS:              device.OS.String(),
		DestURL:         res.DestURL,
		IOSDeepLink:     res.IOSDeepLink,
		AndroidDeepLink: res.AndroidDeepLink,
		IOSFallback:     res.IOSFallbackURL,
		AndroidFallback: res.AndroidFallbackURL,
		AndroidPackage:  res.AndroidPackage,
	}

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return errs.Internal("").WithCause(err)
	}

	title := res.OGTitle
	if title == "" {
		title = "Open Link"
	}
	desc := res.OGDescription
	if desc == "" {
		desc = "Tap the button below to open this link in the app."
	}

	data := deepviewData{
		Title:       title,
		Description: desc,
		OGImage:     res.OGImage,
		DestURL:     res.DestURL,
		ConfigJSON:  string(cfgJSON),
	}

	var buf bytes.Buffer
	if err := deepviewTmpl.Execute(&buf, data); err != nil {
		return errs.Internal("").WithCause(err)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "private, no-store")
	return c.Send(buf.Bytes())
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound),
		errors.Is(err, ErrInactive),
		errors.Is(err, ErrExpired),
		errors.Is(err, ErrNotStarted),
		errors.Is(err, ErrClickLimitReached):
		return errs.NotFound("link", "")
	default:
		return errs.Internal("").WithCause(err)
	}
}
