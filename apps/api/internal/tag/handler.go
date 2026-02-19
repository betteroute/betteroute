package tag

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
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
	tags := r.Group("/tags")
	tags.Get("/", h.List)
	tags.Get("/:id", h.Get)
	tags.Post("/", h.Create)
	tags.Patch("/:id", h.Update)
	tags.Delete("/:id", h.Delete)
}

// RegisterLinkRoutes mounts tag-link association routes.
// Called separately because these live under /links/:id/tags.
func (h *Handler) RegisterLinkRoutes(r fiber.Router) {
	r.Get("/:id/tags", h.ListByLink)
	r.Post("/:id/tags", h.AddToLink)
	r.Delete("/:id/tags/:tag_id", h.RemoveFromLink)
}

// List returns all tags for a workspace.
//
// @Summary     List tags
// @Description Returns all tags for a workspace, ordered by name.
// @Tags        tags
// @Produce     json
// @Param       workspace_id query string true "Workspace ID"
// @Success     200 {array}  Tag
// @Failure     400 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/tags [get]
func (h *Handler) List(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	tags, err := h.svc.List(c.Context(), wsID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(tags)
}

// Get returns a single tag by ID.
//
// @Summary     Get tag
// @Description Returns a single tag by ID within a workspace.
// @Tags        tags
// @Produce     json
// @Param       id           path  string true "Tag ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     200 {object} Tag
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/tags/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	t, err := h.svc.Get(c.Context(), c.Params("id"), wsID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(t)
}

// Create creates a new tag.
//
// @Summary     Create tag
// @Description Creates a new tag in the workspace.
// @Tags        tags
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Tag input"
// @Success     201  {object} Tag
// @Failure     400  {object} errs.Error
// @Failure     409  {object} errs.Error "Tag name already exists"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/tags [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	t, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return h.mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(t)
}

// Update partially updates a tag.
//
// @Summary     Update tag
// @Description Partially updates a tag. Only provided fields are changed.
// @Tags        tags
// @Accept      json
// @Produce     json
// @Param       id           path  string      true "Tag ID"
// @Param       workspace_id query string      true "Workspace ID"
// @Param       body         body  UpdateInput true "Fields to update"
// @Success     200 {object} Tag
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     409 {object} errs.Error "Tag name already exists"
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/tags/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	t, err := h.svc.Update(c.Context(), c.Params("id"), wsID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(t)
}

// Delete soft-deletes a tag.
//
// @Summary     Delete tag
// @Description Soft-deletes a tag. Removes it from all links.
// @Tags        tags
// @Param       id           path  string true "Tag ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     204          "No Content"
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/tags/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	if err := h.svc.Delete(c.Context(), c.Params("id"), wsID); err != nil {
		return h.mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListByLink returns all tags for a link.
//
// @Summary     List tags for link
// @Description Returns all tags associated with a link.
// @Tags        tags
// @Produce     json
// @Param       id path string true "Link ID"
// @Success     200 {array} Tag
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id}/tags [get]
func (h *Handler) ListByLink(c fiber.Ctx) error {
	tags, err := h.svc.ListByLink(c.Context(), c.Params("id"))
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(tags)
}

// AddToLink associates a tag with a link.
//
// @Summary     Add tag to link
// @Description Associates a tag with a link. Idempotent — no error if already associated.
// @Tags        tags
// @Accept      json
// @Param       id   path string       true "Link ID"
// @Param       body body AddToLinkInput true "Tag to associate"
// @Success     204  "No Content"
// @Failure     400 {object} errs.Error
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id}/tags [post]
func (h *Handler) AddToLink(c fiber.Ctx) error {
	var input AddToLinkInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.AddToLink(c.Context(), c.Params("id"), input.TagID); err != nil {
		return h.mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RemoveFromLink removes a tag from a link.
//
// @Summary     Remove tag from link
// @Description Removes a tag association from a link.
// @Tags        tags
// @Param       id     path string true "Link ID"
// @Param       tag_id path string true "Tag ID"
// @Success     204    "No Content"
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id}/tags/{tag_id} [delete]
func (h *Handler) RemoveFromLink(c fiber.Ctx) error {
	if err := h.svc.RemoveFromLink(c.Context(), c.Params("id"), c.Params("tag_id")); err != nil {
		return h.mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("Tag", "")
	case errors.Is(err, ErrNameTaken):
		return errs.Conflict("tag name already exists in this workspace")
	default:
		return errs.Internal("").WithCause(err)
	}
}
