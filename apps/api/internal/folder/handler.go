package folder

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
	r.Get("/", h.List)
	r.Get("/:id", h.Get)
	r.Post("/", h.Create)
	r.Patch("/:id", h.Update)
	r.Delete("/:id", h.Delete)
}

// @Summary     List folders
// @Description Returns all folders for a workspace, ordered by position.
// @Tags        folders
// @Produce     json
// @Success     200 {array}  Folder
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/folders [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeFoldersRead); err != nil {
		return err
	}

	folders, err := h.svc.List(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(folders)
}

// @Summary     Get folder
// @Description Returns a single folder by ID within a workspace.
// @Tags        folders
// @Produce     json
// @Param       id path string true "Folder ID"
// @Success     200 {object} Folder
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/folders/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeFoldersRead); err != nil {
		return err
	}

	f, err := h.svc.Get(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(f)
}

// @Summary     Create folder
// @Description Creates a new folder in the workspace.
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Folder input"
// @Success     201  {object} Folder
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "Quota exceeded"
// @Failure     409  {object} errs.Error "Folder name already exists"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/folders [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeFoldersWrite); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaFolders, 1); err != nil {
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
	f, err := h.svc.Create(ctx, rbac.FromContext(ctx).WorkspaceID, userID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(f)
}

// @Summary     Update folder
// @Description Partially updates a folder. Only provided fields are changed.
// @Tags        folders
// @Accept      json
// @Produce     json
// @Param       id   path string      true "Folder ID"
// @Param       body body UpdateInput  true "Fields to update"
// @Success     200 {object} Folder
// @Failure     400 {object} errs.Error
// @Failure     404 {object} errs.Error
// @Failure     409 {object} errs.Error "Folder name already exists"
// @Failure     422 {object} errs.Error "Validation failed"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/folders/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeFoldersWrite); err != nil {
		return err
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	f, err := h.svc.Update(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(f)
}

// @Summary     Delete folder
// @Description Soft-deletes a folder. Links in the folder become unfiled.
// @Tags        folders
// @Param       id path string true "Folder ID"
// @Success     204 "No Content"
// @Failure     404 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/folders/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeFoldersWrite); err != nil {
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
		return errs.NotFound("folder", "")
	case errors.Is(err, ErrNameTaken):
		return errs.Conflict("folder name already exists in this workspace")
	default:
		return errs.Internal("").WithCause(err)
	}
}
