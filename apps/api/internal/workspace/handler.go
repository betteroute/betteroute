package workspace

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/apikey"
	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/guard"
	"github.com/execrc/betteroute/internal/rbac"
	"github.com/execrc/betteroute/internal/validate"
)

// Handler handles workspace HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new workspace handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts workspace routes on the given router.
// workspaceMW resolves slug → rbac.Context, entitlementMW resolves quota/feature caps.
func (h *Handler) Register(r fiber.Router, workspaceMW, entitlementMW fiber.Handler) {
	ws := r.Group("/workspaces")
	ws.Get("/", h.List)
	ws.Post("/", h.Create)
	ws.Post("/accept-invitation", h.AcceptInvitation)

	slug := ws.Group("/:slug", workspaceMW, entitlementMW)
	slug.Get("/", h.Get)
	slug.Patch("/", h.Update)
	slug.Delete("/", h.Delete)

	members := slug.Group("/members")
	members.Get("/", h.ListMembers)
	members.Patch("/:userID", h.UpdateMember)
	members.Delete("/:userID", h.RemoveMember)

	invs := slug.Group("/invitations")
	invs.Get("/", h.ListInvitations)
	invs.Post("/", h.Invite)
	invs.Delete("/:id", h.CancelInvitation)
}

// @Summary     List workspaces
// @Description Returns all workspaces the authenticated user is a member of, with their role in each.
// @Tags        workspaces
// @Produce     json
// @Success     200 {array}  WithRole
// @Failure     401 {object} errs.Error
// @Failure     403 {object} errs.Error "API keys cannot list workspaces"
// @Failure     500 {object} errs.Error
// @Router      /api/v1/workspaces [get]
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	if apikey.FromContext(ctx) != nil {
		return errs.Forbidden("api keys cannot list workspaces")
	}

	workspaces, err := h.svc.List(ctx, auth.FromContext(ctx).User.ID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(workspaces)
}

// @Summary     Create workspace
// @Description Creates a new workspace and adds the authenticated user as owner.
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       body body     CreateInput true "Workspace input"
// @Success     201  {object} WithRole
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error "API keys cannot create workspaces"
// @Failure     409  {object} errs.Error "Slug already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces [post]
func (h *Handler) Create(c fiber.Ctx) error {
	ctx := c.Context()
	if apikey.FromContext(ctx) != nil {
		return errs.Forbidden("api keys cannot create workspaces")
	}

	var input CreateInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	ws, err := h.svc.Create(ctx, auth.FromContext(ctx).User.ID, input)
	if err != nil {
		return mapError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(WithRole{Workspace: ws, Role: rbac.Owner})
}

// @Summary     Get workspace
// @Description Returns a workspace by slug with the caller's role.
// @Tags        workspaces
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {object} WithRole
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeWorkspaceRead); err != nil {
		return err
	}

	rctx := rbac.FromContext(ctx)
	ws, err := h.svc.Get(ctx, rctx.WorkspaceID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(WithRole{Workspace: ws, Role: rctx.Role})
}

// @Summary     Update workspace
// @Description Partially updates a workspace. Requires Admin role.
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       body body UpdateInput  true "Fields to update"
// @Success     200  {object} WithRole
// @Failure     400  {object} errs.Error
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     409  {object} errs.Error "Slug already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	var input UpdateInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	rctx := rbac.FromContext(ctx)
	ws, err := h.svc.Update(ctx, rctx.WorkspaceID, input)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(WithRole{Workspace: ws, Role: rctx.Role})
}

// @Summary     Delete workspace
// @Description Soft-deletes a workspace. Requires Owner role.
// @Tags        workspaces
// @Param       slug path string true "Workspace slug"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Owner); err != nil {
		return err
	}

	if err := h.svc.Delete(ctx, rbac.FromContext(ctx).WorkspaceID); err != nil {
		return mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     List members
// @Description Returns all members of a workspace with their roles.
// @Tags        workspaces
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  Member
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/members [get]
func (h *Handler) ListMembers(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Scope(ctx, rbac.ScopeWorkspaceRead); err != nil {
		return err
	}

	members, err := h.svc.ListMembers(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(members)
}

// @Summary     Update member role
// @Description Changes a workspace member's role. Requires Admin role. Cannot demote the last owner.
// @Tags        workspaces
// @Accept      json
// @Param       slug   path string           true "Workspace slug"
// @Param       userID path string           true "Target user ID"
// @Param       body   body UpdateMemberInput true "New role"
// @Success     204    "No Content"
// @Failure     400    {object} errs.Error "Cannot remove last owner"
// @Failure     403    {object} errs.Error
// @Failure     404    {object} errs.Error
// @Failure     422    {object} errs.Error "Validation failed"
// @Failure     500    {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/members/{userID} [patch]
func (h *Handler) UpdateMember(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	var input UpdateMemberInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.UpdateMember(ctx, rbac.FromContext(ctx).WorkspaceID, c.Params("userID"), input); err != nil {
		return mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Remove member
// @Description Removes a user from the workspace. Any member can remove themselves (leave). Removing others requires Admin role. Cannot remove the last owner.
// @Tags        workspaces
// @Param       slug   path string true "Workspace slug"
// @Param       userID path string true "Target user ID"
// @Success     204    "No Content"
// @Failure     400    {object} errs.Error "Cannot remove last owner"
// @Failure     403    {object} errs.Error
// @Failure     404    {object} errs.Error
// @Failure     500    {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/members/{userID} [delete]
func (h *Handler) RemoveMember(c fiber.Ctx) error {
	ctx := c.Context()
	targetUserID := c.Params("userID")
	user := auth.FromContext(ctx).User

	// Any member can remove themselves (leave). Removing others requires Admin+.
	if targetUserID != user.ID {
		if err := guard.Role(ctx, rbac.Admin); err != nil {
			return err
		}
	}

	if err := h.svc.RemoveMember(ctx, rbac.FromContext(ctx).WorkspaceID, targetUserID); err != nil {
		return mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Invite member
// @Description Creates a new workspace invitation and sends an email. Requires Admin role.
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       slug path string      true "Workspace slug"
// @Param       body body InviteInput  true "Invitation payload"
// @Success     201  {object} Invitation
// @Failure     400  {object} errs.Error
// @Failure     402  {object} errs.Error "Member quota exceeded"
// @Failure     403  {object} errs.Error
// @Failure     409  {object} errs.Error "Already invited or already a member"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/invitations [post]
func (h *Handler) Invite(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}
	if err := guard.Quota(ctx, entitlement.QuotaMembers, 1); err != nil {
		return err
	}

	var input InviteInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user := auth.FromContext(ctx).User
	inv, err := h.svc.Invite(ctx, rbac.FromContext(ctx).WorkspaceID, user.ID, user.Name, input)
	if err != nil {
		return mapError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(inv)
}

// @Summary     List invitations
// @Description Returns all pending invitations for the workspace. Requires Admin role.
// @Tags        workspaces
// @Produce     json
// @Param       slug path string true "Workspace slug"
// @Success     200  {array}  Invitation
// @Failure     403  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/invitations [get]
func (h *Handler) ListInvitations(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	invitations, err := h.svc.ListInvitations(ctx, rbac.FromContext(ctx).WorkspaceID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(invitations)
}

// @Summary     Cancel invitation
// @Description Deletes a pending workspace invitation. Requires Admin role.
// @Tags        workspaces
// @Param       slug path string true "Workspace slug"
// @Param       id   path string true "Invitation ID"
// @Success     204  "No Content"
// @Failure     403  {object} errs.Error
// @Failure     404  {object} errs.Error
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/{slug}/invitations/{id} [delete]
func (h *Handler) CancelInvitation(c fiber.Ctx) error {
	ctx := c.Context()
	if err := guard.Role(ctx, rbac.Admin); err != nil {
		return err
	}

	if err := h.svc.CancelInvitation(ctx, rbac.FromContext(ctx).WorkspaceID, c.Params("id")); err != nil {
		return mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Accept invitation
// @Description Accepts a workspace invitation using a token. The authenticated user must match the invited email.
// @Tags        workspaces
// @Accept      json
// @Produce     json
// @Param       body body     AcceptInvitationInput true "Invitation token"
// @Success     200  {object} WithRole
// @Failure     400  {object} errs.Error "Token invalid or expired"
// @Failure     403  {object} errs.Error "Email mismatch or API key auth"
// @Failure     409  {object} errs.Error "Already a member"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/workspaces/accept-invitation [post]
func (h *Handler) AcceptInvitation(c fiber.Ctx) error {
	ctx := c.Context()
	if apikey.FromContext(ctx) != nil {
		return errs.Forbidden("api keys cannot accept invitations")
	}

	var input AcceptInvitationInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user := auth.FromContext(ctx).User
	ws, err := h.svc.AcceptInvitation(ctx, user.ID, user.Email, input)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(ws)
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("workspace", "")
	case errors.Is(err, ErrSlugTaken):
		return errs.Conflict("slug already in use")
	case errors.Is(err, ErrNotMember):
		return errs.Forbidden("")
	case errors.Is(err, ErrAlreadyMember):
		return errs.Conflict("user is already a member")
	case errors.Is(err, ErrCannotRemoveOwner):
		return errs.BadRequest("cannot remove the last owner")
	case errors.Is(err, ErrInvalidSlug):
		return errs.BadRequest("workspace name produced an invalid slug")
	case errors.Is(err, ErrTokenInvalid):
		return errs.BadRequest("invitation token is invalid or expired")
	case errors.Is(err, ErrInviteMismatch):
		return errs.Forbidden("invitation is for a different email address")
	case errors.Is(err, ErrAlreadyInvited):
		return errs.Conflict("a pending invitation already exists for this email")
	default:
		return errs.Internal("").WithCause(err)
	}
}
