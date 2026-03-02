package domain

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/guard"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles domain HTTP requests.
type Handler struct {
	svc             *Service
	platformDomains []string
}

// NewHandler creates a new domain handler.
func NewHandler(svc *Service, platformDomains []string) *Handler {
	return &Handler{svc: svc, platformDomains: platformDomains}
}

// Register mounts domain CRUD and verification routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/", h.List)
	r.Get("/platform", h.ListPlatform)
	r.Get("/:id", h.Get)
	r.Post("/", h.Create)
	r.Patch("/:id", h.Update)
	r.Delete("/:id", h.Delete)
	r.Post("/:id/verify", h.Verify)
}

// @Summary     List domains
// @Description Returns all custom domains for a workspace.
// @Tags        domains
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  Domain
// @Failure     403  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeDomainsRead); err != nil {
		return err
	}

	domains, err := h.svc.List(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(domains)
}

// @Summary     List platform domains
// @Description Returns the platform-owned short link domains available to all users.
// @Tags        domains
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array} string
// @Router      /api/v1/workspaces/{slug}/domains/platform [get]
func (h *Handler) ListPlatform(c fiber.Ctx) error {
	return c.JSON(h.platformDomains)
}

// @Summary     Get domain
// @Description Returns a single domain by ID within a workspace.
// @Tags        domains
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Domain ID"
// @Success     200  {object} Domain
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeDomainsRead); err != nil {
		return err
	}

	d, err := h.svc.Get(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(d)
}

// @Summary     Add domain
// @Description Adds a custom domain to the workspace. Returns DNS setup instructions.
// @Tags        domains
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       body body  CreateInput true "Domain input"
// @Success     201  {object} Domain
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "Quota exceeded"
// @Failure     403  {object} errs.Error
// @Failure     409  {object} errs.Error "Hostname already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeDomainsWrite); err != nil {
		return err
	}
	if err := guard.Feature(ctx, entitlement.FeatureCustomDomains); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaDomains, 1); err != nil {
		return err
	}

	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	userID := auth.FromContext(ctx).User.ID
	d, err := h.svc.Create(ctx, rbac.FromContext(ctx).WorkspaceID, userID, input)
	if err != nil {
		return mapError(err)
	}

	// Return the domain with DNS setup instructions.
	type response struct {
		Domain
		DNS DNSSetup `json:"dns"`
	}

	return c.Status(fiber.StatusCreated).JSON(response{
		Domain: *d,
		DNS:    d.DNSInstructions(),
	})
}

// @Summary     Update domain
// @Description Partially updates a domain. Only provided fields are changed.
// @Tags        domains
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       id   path string      true "Domain ID"
// @Param       body body  UpdateInput true "Fields to update"
// @Success     200  {object} Domain
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeDomainsWrite); err != nil {
		return err
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	d, err := h.svc.Update(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID, input)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(d)
}

// @Summary     Delete domain
// @Description Soft-deletes a domain. Links using this domain fall back to the platform default.
// @Tags        domains
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Domain ID"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeDomainsWrite); err != nil {
		return err
	}

	if err := h.svc.Delete(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID); err != nil {
		return mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Verify domain
// @Description Checks DNS TXT record to verify domain ownership and activates the domain.
// @Tags        domains
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Domain ID"
// @Success     200  {object} Domain
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     409  {object} errs.Error "Already verified"
// @Failure     422  {object} errs.Error "DNS verification failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/domains/{id}/verify [post]
func (h *Handler) Verify(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeDomainsWrite); err != nil {
		return err
	}

	d, err := h.svc.Verify(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(d)
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("domain", "")
	case errors.Is(err, ErrHostnameTaken):
		return errs.Conflict("hostname already in use")
	case errors.Is(err, ErrAlreadyVerified):
		return errs.Conflict("domain is already verified")
	case errors.Is(err, ErrDNSNotFound):
		return errs.Validation([]errs.FieldError{{
			Field:   "hostname",
			Message: "no TXT record found on _betteroute." + " — ensure the record is set and DNS has propagated",
		}})
	case errors.Is(err, ErrDNSMismatch):
		return errs.Validation([]errs.FieldError{{
			Field:   "hostname",
			Message: "TXT record found but value does not match the expected verification token",
		}})
	default:
		return errs.Internal("").WithCause(err)
	}
}
