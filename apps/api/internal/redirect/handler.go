package redirect

import (
	"bytes"
	_ "embed"
	"errors"
	"html/template"

	"github.com/gofiber/fiber/v3"
	useragent "github.com/medama-io/go-useragent"

	"github.com/execrc/betteroute/internal/errs"
)

//go:embed templates/og.html
var ogHTML string

// ogTmpl is the parsed OG template, initialized once at package load.
var ogTmpl = template.Must(template.New("og").Parse(ogHTML))

// Handler handles short code redirect requests.
type Handler struct {
	svc *Service
	ua  *useragent.Parser
}

// NewHandler creates a new redirect handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
		ua:  useragent.NewParser(),
	}
}

// Register mounts the catch-all redirect route.
// Must be registered LAST to avoid catching /api/v1 and /health.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/:code", h.Redirect)
}

// Redirect resolves a short code and issues a 302 redirect.
// Social crawlers receive an HTML page with OG meta tags instead.
//
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

	res, err := h.svc.Resolve(c.Context(), code)
	if err != nil {
		return h.mapError(err)
	}

	// Serve OG HTML to social crawlers for rich preview cards.
	if res.HasOG() && h.ua.Parse(c.Get("User-Agent")).IsBot() {
		return h.serveOG(c, res)
	}

	return c.Redirect().Status(fiber.StatusFound).To(res.DestURL)
}

// serveOG renders the OG template and returns it as HTML.
func (h *Handler) serveOG(c fiber.Ctx, res *Resolution) error {
	var buf bytes.Buffer
	if err := ogTmpl.Execute(&buf, res); err != nil {
		return errs.Internal("").WithCause(err)
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound),
		errors.Is(err, ErrInactive),
		errors.Is(err, ErrExpired),
		errors.Is(err, ErrNotStarted),
		errors.Is(err, ErrClickLimitReached):
		return errs.NotFound("Link", "")
	default:
		return errs.Internal("").WithCause(err)
	}
}
