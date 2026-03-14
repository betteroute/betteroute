package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	useragent "github.com/medama-io/go-useragent"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/validate"
)

// CookieName is the session cookie name, shared with the auth middleware.
const CookieName = "session"

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc           *Service
	secureCookies bool
	ua            *useragent.Parser
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service, secureCookies bool) *Handler {
	return &Handler{
		svc:           svc,
		secureCookies: secureCookies,
		ua:            useragent.NewParser(),
	}
}

// Register mounts all auth routes onto the router.
func (h *Handler) Register(r fiber.Router, authMW fiber.Handler) {
	a := r.Group("/auth")

	// Public routes — no session required.
	a.Post("/magic-link", h.SendMagicLink)
	a.Post("/verify-magic-link", h.VerifyMagicLink)
	a.Get("/oauth/:provider", h.OAuthRedirect)
	a.Get("/oauth/:provider/callback", h.OAuthCallback)

	// Protected routes — require a valid session or API key.
	protected := a.Group("", authMW)
	protected.Post("/logout", h.Logout)
	protected.Get("/me", h.Me)
	protected.Patch("/me", h.UpdateMe)
}

// @Summary     Request Magic Link
// @Description Sends a one-time login link to the user's email.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     MagicLinkInput true "Email address"
// @Success     204  "No Content"
// @Failure     400  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     429  {object} errs.Error "Rate limited"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/magic-link [post]
func (h *Handler) SendMagicLink(c fiber.Ctx) error {
	var input MagicLinkInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.SendMagicLink(c.Context(), input); err != nil {
		return mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Verify Magic Link
// @Description Authenticates a user via a magic link token and sets a session cookie.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     VerifyMagicLinkInput true "Magic link token"
// @Success     200  {object} User
// @Failure     400  {object} errs.Error "Token invalid or expired"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/verify-magic-link [post]
func (h *Handler) VerifyMagicLink(c fiber.Ctx) error {
	var input VerifyMagicLinkInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user, sess, err := h.svc.VerifyMagicLink(c.Context(), input, h.sessionMeta(c))
	if err != nil {
		return mapError(err)
	}

	h.setSessionCookie(c, sess)
	return c.JSON(user)
}

// @Summary     Log out
// @Description Invalidates the current session and clears the session cookie.
// @Tags        auth
// @Success     204 "No Content"
// @Failure     400 {object} errs.Error "No active session (API key auth)"
// @Failure     401 {object} errs.Error
// @Failure     500 {object} errs.Error
// @Router      /api/v1/auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	session := FromContext(c.Context()).Session
	if session == nil {
		return errs.BadRequest("no active session")
	}
	if err := h.svc.Logout(c.Context(), session.ID); err != nil {
		return mapError(err)
	}
	h.clearSessionCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Get current user
// @Description Returns the authenticated user's profile.
// @Tags        auth
// @Produce     json
// @Success     200 {object} User
// @Failure     401 {object} errs.Error
// @Router      /api/v1/auth/me [get]
func (h *Handler) Me(c fiber.Ctx) error {
	// User is already loaded by Authenticate middleware.
	return c.JSON(FromContext(c.Context()).User)
}

// @Summary     Update current user
// @Description Partially updates the authenticated user's profile.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     UpdateProfileInput true "Fields to update"
// @Success     200  {object} User
// @Failure     400  {object} errs.Error
// @Failure     401  {object} errs.Error
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/me [patch]
func (h *Handler) UpdateMe(c fiber.Ctx) error {
	var input UpdateProfileInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user, err := h.svc.UpdateProfile(c.Context(), FromContext(c.Context()).User.ID, input)
	if err != nil {
		return mapError(err)
	}

	return c.JSON(user)
}

// @Summary     OAuth redirect
// @Description Redirects to the OAuth provider's authorization page.
// @Tags        auth
// @Param       provider path string true "OAuth provider" Enums(google, github)
// @Success     302      "Redirect to provider"
// @Failure     400      {object} errs.Error "Provider not configured"
// @Failure     500      {object} errs.Error
// @Router      /api/v1/auth/oauth/{provider} [get]
func (h *Handler) OAuthRedirect(c fiber.Ctx) error {
	provider := c.Params("provider")

	// Generate random state and store it in a short-lived cookie for CSRF protection.
	state, _, err := generateToken()
	if err != nil {
		return errs.Internal("").WithCause(err)
	}
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300, // 5 minutes
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
	})

	url, err := h.svc.OAuthURL(provider, state)
	if err != nil {
		return mapError(err)
	}
	return c.Redirect().To(url)
}

// @Summary     OAuth callback
// @Description Handles the provider redirect, creates or links the account, and sets a session cookie.
// @Tags        auth
// @Param       provider path  string true "OAuth provider" Enums(google, github)
// @Param       code     query string true "Authorization code"
// @Param       state    query string true "CSRF state"
// @Success     302      "Redirect to frontend callback"
// @Failure     400      {object} errs.Error "Invalid state or missing code"
// @Failure     500      {object} errs.Error
// @Router      /api/v1/auth/oauth/{provider}/callback [get]
func (h *Handler) OAuthCallback(c fiber.Ctx) error {
	provider := c.Params("provider")
	code := c.Query("code")
	state := c.Query("state")

	// Verify CSRF state matches what we set in the cookie.
	if c.Cookies("oauth_state") != state {
		return errs.BadRequest("invalid oauth state")
	}
	c.ClearCookie("oauth_state")

	if code == "" {
		return errs.BadRequest("missing oauth code")
	}

	_, sess, err := h.svc.OAuthCallback(c.Context(), provider, code, h.sessionMeta(c))
	if err != nil {
		return mapError(err)
	}

	h.setSessionCookie(c, sess)

	// Redirect to frontend callback — frontend decides routing (onboarding, dashboard, etc.).
	return c.Redirect().To(h.svc.cfg.WebURL + "/auth/callback")
}

// mapError maps domain errors to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, ErrEmailTaken):
		return errs.Conflict("email already in use")
	case errors.Is(err, ErrInvalidCredential):
		return errs.Unauthorized("invalid credentials")
	case errors.Is(err, ErrAccountSuspended):
		return errs.Forbidden("account is suspended")
	case errors.Is(err, ErrTokenInvalid):
		return errs.BadRequest("token is invalid or expired")
	case errors.Is(err, ErrRateLimited):
		return errs.TooManyRequests(60)
	case errors.Is(err, ErrOAuthUnavailable):
		return errs.BadRequest("oauth provider not available")
	case errors.Is(err, ErrSessionNotFound):
		return errs.Unauthorized("")
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("user", "")
	case errors.Is(err, ErrEmailUnverified):
		return errs.BadRequest("provider email is not verified")
	default:
		return errs.Internal("").WithCause(err)
	}
}

func (h *Handler) setSessionCookie(c fiber.Ctx, sess *Session) {
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    sess.Token,
		Expires:  sess.ExpiresAt,
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
}

func (h *Handler) clearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
	})
}

// sessionMeta extracts server-side HTTP metadata for session creation.
func (h *Handler) sessionMeta(c fiber.Ctx) SessionMeta {
	rawUA := c.Get(fiber.HeaderUserAgent)
	parsed := h.ua.Parse(rawUA)

	browser := string(parsed.Browser())
	if browser == "" {
		browser = "Unknown Browser"
	}

	version := parsed.BrowserVersionMajor()
	if version != "" {
		browser += " " + version
	}

	uaStr := browser
	osName := string(parsed.OS())
	if osName != "" {
		uaStr += " (" + osName + ")"
	}

	return SessionMeta{
		IPAddress: c.IP(),
		UserAgent: uaStr,
	}
}
