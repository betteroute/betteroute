package apikey

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/guard"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles API key HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new API key handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts API key routes on the given router.
// No scope guards — API key management is dashboard-only (session auth required).
func (h *Handler) Register(r fiber.Router) {
	r.Get("/", h.List)
	r.Get("/:id", h.Get)
	r.Post("/", h.Create)
	r.Delete("/:id", h.Delete)
}

// @Summary     List API keys
// @Description Returns all active API keys for the workspace. Requires Admin role.
// @Tags        api-keys
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  APIKey
// @Failure     403  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/api-keys [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	keys, err := h.svc.List(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(keys)
}

// @Summary     Get API key
// @Description Returns a single API key by ID. Requires Admin role.
// @Tags        api-keys
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "API key ID"
// @Success     200  {object} APIKey
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/api-keys/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	key, err := h.svc.Get(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(key)
}

// @Summary     Create API key
// @Description Creates a new API key for the workspace. The raw key is returned only once. Requires Admin role.
// @Tags        api-keys
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       body body CreateInput  true "API key input"
// @Success     201  {object} APIKey
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "API key quota exceeded"
// @Failure     403  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/api-keys [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaAPIKeys, 1); err != nil {
		return err
	}

	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	result, err := h.svc.Create(
		ctx,
		rbac.FromContext(ctx).WorkspaceID,
		auth.FromContext(ctx).User.ID,
		input,
	)
	if err != nil {
		return mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// @Summary     Delete API key
// @Description Permanently removes an API key. Requires Admin role.
// @Tags        api-keys
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "API key ID"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/api-keys/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	if err := h.svc.Delete(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID); err != nil {
		return mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("API key", "")
	case errors.Is(err, ErrInvalidScope):
		return errs.BadRequest("invalid scope")
	case errors.Is(err, ErrScopesRequired):
		return errs.BadRequest("restricted permission requires at least one scope")
	default:
		return errs.Internal("").WithCause(err)
	}
}
