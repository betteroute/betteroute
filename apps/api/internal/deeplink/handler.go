package deeplink

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

// Handler handles workspace app and platform app HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new deeplink handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts workspace app CRUD and platform app list routes.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/platform-apps", h.ListPlatformApps)

	r.Get("/", h.ListWorkspaceApps)
	r.Get("/:id", h.GetWorkspaceApp)
	r.Post("/", h.CreateWorkspaceApp)
	r.Patch("/:id", h.UpdateWorkspaceApp)
	r.Delete("/:id", h.DeleteWorkspaceApp)
}

// @Summary     List platform apps
// @Description Returns the catalog of platform-supported apps for deep linking.
// @Tags        apps
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  PlatformApp
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps/platform-apps [get]
func (h *Handler) ListPlatformApps(c fiber.Ctx) error {
	apps, err := h.svc.ListPlatformApps(c.Context())
	if err != nil {
		return mapError(err)
	}
	return c.JSON(apps)
}

// @Summary     List workspace apps
// @Description Returns all custom apps registered by the workspace.
// @Tags        apps
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  WorkspaceApp
// @Failure     403  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps [get]
func (h *Handler) ListWorkspaceApps(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeAppsRead); err != nil {
		return err
	}

	apps, err := h.svc.ListWorkspaceApps(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(apps)
}

// @Summary     Get workspace app
// @Description Returns a single workspace app by ID.
// @Tags        apps
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Workspace App ID"
// @Success     200  {object} WorkspaceApp
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps/{id} [get]
func (h *Handler) GetWorkspaceApp(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeAppsRead); err != nil {
		return err
	}

	wa, err := h.svc.GetWorkspaceApp(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(wa)
}

// @Summary     Create workspace app
// @Description Registers a custom iOS or Android app for deep linking.
// @Tags        apps
// @Accept      json
// @Produce     json
// @Param       slug path string                  true "Workspace slug"
// @Param       body body  CreateWorkspaceAppInput true "App input"
// @Success     201  {object} WorkspaceApp
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps [post]
func (h *Handler) CreateWorkspaceApp(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeAppsWrite); err != nil {
		return err
	}
	if err := guard.Feature(ctx, entitlement.FeatureDeepLinking); err != nil {
		return err
	}

	var input CreateWorkspaceAppInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	userID := auth.FromContext(ctx).User.ID
	wa, err := h.svc.CreateWorkspaceApp(ctx, rbac.FromContext(ctx).WorkspaceID, userID, input)
	if err != nil {
		return mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(wa)
}

// @Summary     Update workspace app
// @Description Partially updates a workspace app. Only provided fields are changed.
// @Tags        apps
// @Accept      json
// @Produce     json
// @Param       slug path string                  true "Workspace slug"
// @Param       id   path string                  true "Workspace App ID"
// @Param       body body  UpdateWorkspaceAppInput true "Fields to update"
// @Success     200  {object} WorkspaceApp
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps/{id} [patch]
func (h *Handler) UpdateWorkspaceApp(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeAppsWrite); err != nil {
		return err
	}
	if err := guard.Feature(ctx, entitlement.FeatureDeepLinking); err != nil {
		return err
	}

	var input UpdateWorkspaceAppInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	wa, err := h.svc.UpdateWorkspaceApp(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID, input)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(wa)
}

// @Summary     Delete workspace app
// @Description Soft-deletes a workspace app.
// @Tags        apps
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Workspace App ID"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/apps/{id} [delete]
func (h *Handler) DeleteWorkspaceApp(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeAppsWrite); err != nil {
		return err
	}

	if err := h.svc.DeleteWorkspaceApp(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID); err != nil {
		return mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapError(err error) error {
	switch {
	case errors.Is(err, ErrWorkspaceAppNotFound):
		return errs.NotFound("workspace app", "")
	default:
		return errs.Internal("").WithCause(err)
	}
}
