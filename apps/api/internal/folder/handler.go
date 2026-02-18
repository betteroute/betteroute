package folder

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/patch"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles folder HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new folder handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts folder CRUD routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	folders := r.Group("/folders")
	folders.Get("/", h.List)
	folders.Get("/:id", h.Get)
	folders.Post("/", h.Create)
	folders.Patch("/:id", h.Update)
	folders.Delete("/:id", h.Delete)
}

// List returns all folders for a workspace.
//
// @Summary     List folders
// @Description Returns all folders for a workspace, ordered by position.
// @Tags        folders
// @Produce     json
// @Param       workspace_id query string true "Workspace ID"
// @Success     200 {array}  Folder
// @Failure     400 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/folders [get]
func (h *Handler) List(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	folders, err := h.svc.List(c.Context(), wsID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(folders)
}

// Get returns a single folder by ID.
//
// @Summary     Get folder
// @Description Returns a single folder by ID within a workspace.
// @Tags        folders
// @Produce     json
// @Param       id           path  string true "Folder ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     200 {object} Folder
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/folders/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	f, err := h.svc.Get(c.Context(), c.Params("id"), wsID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(f)
}

// Create creates a new folder.
//
// @Summary     Create folder
// @Description Creates a new folder in the workspace.
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Folder input"
// @Success     201  {object} Folder
// @Failure     400  {object} errs.Error
// @Failure     409  {object} errs.Error "Folder name already exists"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/folders [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	f, err := h.svc.Create(c.Context(), input)
	if err != nil {
		return h.mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(f)
}

// Update partially updates a folder.
//
// @Summary     Update folder
// @Description Partially updates a folder. Only provided fields are changed.
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id           path  string      true "Folder ID"
// @Param       workspace_id query string      true "Workspace ID"
// @Param       body         body  UpdateInput true "Fields to update"
// @Success     200 {object} Folder
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     409 {object} errs.Error "Folder name already exists"
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/folders/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return errs.BadRequest("workspace_id is required")
	}

	input, fields, err := patch.Bind[UpdateInput](c.Body())
	if err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	f, err := h.svc.Update(c.Context(), c.Params("id"), wsID, input, NullableFields{
		Position: fields.Has("position"),
	})
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(f)
}

// Delete soft-deletes a folder.
//
// @Summary     Delete folder
// @Description Soft-deletes a folder. Links in the folder become unfiled.
// @Tags        folders
// @Param       id           path  string true "Folder ID"
// @Param       workspace_id query string true "Workspace ID"
// @Success     204          "No Content"
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/folders/{id} [delete]
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
		return errs.NotFound("Folder", "")
	case errors.Is(err, ErrNameTaken):
		return errs.Conflict("folder name already exists in this workspace")
	default:
		return errs.Internal("").WithCause(err)
	}
}
