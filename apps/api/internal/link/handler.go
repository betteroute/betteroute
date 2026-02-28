package link

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/apikey"
	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/guard"
	"github.com/execrc/betteroute/internal/page"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/tag"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles link HTTP requests.
type Handler struct {
	svc    *Service
	tagSvc *tag.Service
}

// NewHandler creates a new link handler.
func NewHandler(svc *Service, tagSvc *tag.Service) *Handler {
	return &Handler{svc: svc, tagSvc: tagSvc}
}

// Register mounts link CRUD and sub-resource routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/", h.List)
	r.Get("/:id", h.Get)
	r.Post("/", h.Create)
	r.Patch("/:id", h.Update)
	r.Delete("/:id", h.Delete)

	// Tag associations: /workspaces/:slug/links/:id/tags
	r.Get("/:id/tags", h.ListTags)
	r.Post("/:id/tags", h.AddTag)
	r.Delete("/:id/tags/:tag_id", h.RemoveTag)
}

// @Summary     List links
// @Description Returns a paginated list of links for a workspace.
// @Tags        links
// @Produce     json
// @Param       slug     path  string true  "Workspace slug"
// @Param       page     query int    false "Page number"    default(1)
// @Param       per_page query int    false "Items per page" default(20)
// @Success     200 {object} object "Paginated list of links"
// @Failure     403 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeLinksRead); err != nil {
		return err
	}

	pg, perPg := page.Normalize(parseQueryInt(c, "page"), parseQueryInt(c, "per_page"))
	offset := page.Offset(pg, perPg)

	links, total, err := h.svc.List(ctx, rbac.FromContext(ctx).WorkspaceID, perPg, offset)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(page.NewList(links, pg, perPg, total))
}

// @Summary     Get link
// @Description Returns a single link by ID within a workspace.
// @Tags        links
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Link ID"
// @Success     200  {object} Link
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeLinksRead); err != nil {
		return err
	}

	l, err := h.svc.Get(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(l)
}

// @Summary     Create link
// @Description Creates a new short link in the workspace.
// @Tags        links
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       body body  CreateInput true "Link input"
// @Success     201  {object} Link
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "Quota exceeded"
// @Failure     403  {object} errs.Error
// @Failure     409  {object} errs.Error "Short code already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeLinksWrite); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaLinks, 1); err != nil {
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
	createdVia := "web"
	if apikey.FromContext(ctx) != nil {
		createdVia = "api"
	}

	l, err := h.svc.Create(ctx, rbac.FromContext(ctx).WorkspaceID, userID, createdVia, input)
	if err != nil {
		return mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(l)
}

// @Summary     Update link
// @Description Partially updates a link. Only provided fields are changed.
// @Tags        links
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       id   path string      true "Link ID"
// @Param       body body  UpdateInput true "Fields to update"
// @Success     200  {object} Link
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeLinksWrite); err != nil {
		return err
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	l, err := h.svc.Update(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID, input)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(l)
}

// @Summary     Delete link
// @Description Soft-deletes a link. The link is not permanently removed.
// @Tags        links
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Link ID"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeLinksWrite); err != nil {
		return err
	}

	if err := h.svc.Delete(ctx, c.Params("id"), rbac.FromContext(ctx).WorkspaceID); err != nil {
		return mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     List tags for link
// @Description Returns all tags associated with a link.
// @Tags        links
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Link ID"
// @Success     200  {array}  tag.Tag
// @Failure     403  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id}/tags [get]
func (h *Handler) ListTags(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeTagsRead); err != nil {
		return err
	}

	tags, err := h.tagSvc.ListByLink(ctx, c.Params("id"))
	if err != nil {
		return errs.Internal("").WithCause(err)
	}

	return c.JSON(tags)
}

// @Summary     Add tag to link
// @Description Associates a tag with a link. Idempotent — no error if already associated.
// @Tags        links
// @Accept      json
// @Param       slug path string             true "Workspace slug"
// @Param       id   path string             true "Link ID"
// @Param       body body tag.AddToLinkInput  true "Tag to associate"
// @Success     204  "No Content"
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id}/tags [post]
func (h *Handler) AddTag(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeTagsWrite); err != nil {
		return err
	}

	var input tag.AddToLinkInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}

	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.tagSvc.AddToLink(ctx, c.Params("id"), input.TagID); err != nil {
		return errs.Internal("").WithCause(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Remove tag from link
// @Description Removes a tag association from a link.
// @Tags        links
// @Param       slug   path string true "Workspace slug"
// @Param       id     path string true "Link ID"
// @Param       tag_id path string true "Tag ID"
// @Success     204    "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/links/{id}/tags/{tag_id} [delete]
func (h *Handler) RemoveTag(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Member); err != nil {
		return err
	}
	if err := guard.Scope(ctx, rbac.ScopeTagsWrite); err != nil {
		return err
	}

	if err := h.tagSvc.RemoveFromLink(ctx, c.Params("id"), c.Params("tag_id")); err != nil {
		return errs.Internal("").WithCause(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("link", "")
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
