package redirect

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
)

// Handler handles short code redirect requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new redirect handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts the catch-all redirect route.
// Must be registered LAST to avoid catching /api/v1 and /health.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/:code", h.Redirect)
}

// Redirect resolves a short code and issues a 302 redirect.
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

	return c.Redirect().Status(fiber.StatusFound).To(res.DestURL)
}

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound),
		errors.Is(err, ErrInactive),
		errors.Is(err, ErrExpired):
		return errs.NotFound("Link", "")
	default:
		return errs.Internal("").WithCause(err)
	}
}
