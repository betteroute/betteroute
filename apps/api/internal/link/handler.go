package link

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/page"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles link HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new link handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts link CRUD routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	links := r.Group("/links")
	links.Get("/", h.List)
	links.Get("/:id", h.Get)
	links.Post("/", h.Create)
	links.Patch("/:id", h.Update)
	links.Delete("/:id", h.Delete)
}

// List returns paginated links for a workspace.
//
// @Summary     List links
// @Description Returns a paginated list of links for a workspace.
// @Tags        links
// @Accept      json
// @Produce     json
// @Param       workspace_id query    string true  "Workspace ID"
// @Param       page         query    int    false "Page number"     default(1)
// @Param       per_page     query    int    false "Items per page"  default(20)
// @Success     200 {object} object "Paginated list of links"
// @Failure     400 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links [get]
func (h *Handler) List(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	pg, perPg := page.Normalize(parseQueryInt(c, "page"), parseQueryInt(c, "per_page"))
	offset := page.Offset(pg, perPg)

	links, total, err := h.svc.List(c.Context(), wsID, perPg, offset)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(page.NewList(links, pg, perPg, total))
}

// Get returns a single link by ID.
//
// @Summary     Get link
// @Description Returns a single link by ID within a workspace.
// @Tags        links
// @Produce     json
// @Param       id           path  string true "Link ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     200 {object} Link
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	l, err := h.svc.Get(c.Context(), c.Params("id"), wsID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(l)
}

// Create creates a new link.
//
// @Summary     Create link
// @Description Creates a new short link in the workspace.
// @Tags        links
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Link input"
// @Success     201  {object} Link
// @Failure     400  {object} errs.Error
// @Failure     409  {object} errs.Error "Short code already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/links [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	l, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return h.mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(l)
}

// Update partially updates a link.
//
// @Summary     Update link
// @Description Partially updates a link. Only provided fields are changed.
// @Tags        links
// @Accept      json
// @Produce     json
// @Param       id           path  string      true "Link ID"
// @Param       workspace_id query string      true "Workspace ID"
// @Param       body         body  UpdateInput true "Fields to update"
// @Success     200 {object} Link
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id} [patch]
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

	l, err := h.svc.Update(c.Context(), c.Params("id"), wsID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(l)
}

// Delete soft-deletes a link.
//
// @Summary     Delete link
// @Description Soft-deletes a link. The link is not permanently removed.
// @Tags        links
// @Param       id           path  string true "Link ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     204          "No Content"
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/links/{id} [delete]
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

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("Link", "")
	case errors.Is(err, ErrShortCodeTaken):
		return errs.Conflict("short code already in use")
	default:
		return errs.Internal("").WithCause(err)
	}
}

// parseQueryInt reads an integer query param, returning 0 if missing or invalid.
func parseQueryInt(c fiber.Ctx, key string) int {
	v, _ := strconv.Atoi(c.Query(key))
	return v
}
