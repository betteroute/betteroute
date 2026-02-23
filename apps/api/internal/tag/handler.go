package tag

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

// Handler handles tag HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new tag handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts tag CRUD routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/", h.List)
	r.Get("/:id", h.Get)
	r.Post("/", h.Create)
	r.Patch("/:id", h.Update)
	r.Delete("/:id", h.Delete)
}

// @Summary     List tags
// @Description Returns all tags for a workspace, ordered by name.
// @Tags        tags
// @Produce     json
// @Success     200 {array}  Tag
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/tags [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeTagsRead); err != nil {
		return err
	}

	tags, err := h.svc.List(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(tags)
}

// @Summary     Get tag
// @Description Returns a single tag by ID within a workspace.
// @Tags        tags
// @Produce     json
// @Param       id path string true "Tag ID"
// @Success     200 {object} Tag
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/tags/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeTagsRead); err != nil {
		return err
	}

	t, err := h.svc.Get(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(t)
}

// @Summary     Create tag
// @Description Creates a new tag in the workspace.
// @Tags        tags
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Tag input"
// @Success     201  {object} Tag
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "Quota exceeded"
// @Failure     409  {object} errs.Error "Tag name already exists"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/tags [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeTagsWrite); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaTags, 1); err != nil {
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
	t, err := h.svc.Create(ctx, rbac.FromContext(ctx).WorkspaceID, userID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(t)
}

// @Summary     Update tag
// @Description Partially updates a tag. Only provided fields are changed.
// @Tags        tags
// @Accept      json
// @Produce     json
// @Param       id   path string      true "Tag ID"
// @Param       body body UpdateInput  true "Fields to update"
// @Success     200 {object} Tag
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     409 {object} errs.Error "Tag name already exists"
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/tags/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeTagsWrite); err != nil {
		return err
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	t, err := h.svc.Update(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(t)
}

// @Summary     Delete tag
// @Description Soft-deletes a tag. Removes it from all links.
// @Tags        tags
// @Param       id path string true "Tag ID"
// @Success     204 "No Content"
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/tags/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeTagsWrite); err != nil {
		return err
	}

	if err := h.svc.Delete(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID); err != nil {
		return h.mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("tag", "")
	case errors.Is(err, ErrNameTaken):
		return errs.Conflict("tag name already exists in this workspace")
	default:
		return errs.Internal("").WithCause(err)
	}
}
