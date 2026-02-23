package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/validate"
)

// CookieName is the session cookie name, shared with the auth middleware.
const CookieName = "session"

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc           *Service
	secureCookies bool // controls the Secure flag on session cookies
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service, secureCookies bool) *Handler {
	return &Handler{svc: svc, secureCookies: secureCookies}
}

// Register mounts all auth routes onto the router.
func (h *Handler) Register(r fiber.Router, authMW fiber.Handler) {
	a := r.Group("/auth")

	// Public routes — no session required.
	a.Post("/signup", h.SignUp)
	a.Post("/login", h.Login)
	a.Post("/verify-email", h.VerifyEmail)
	a.Post("/resend-verification", h.ResendVerification)
	a.Post("/forgot-password", h.ForgotPassword)
	a.Post("/reset-password", h.ResetPassword)
	a.Get("/oauth/:provider", h.OAuthRedirect)
	a.Get("/oauth/:provider/callback", h.OAuthCallback)

	// Protected routes — require a valid session or API key.
	protected := a.Group("", authMW)
	protected.Post("/logout", h.Logout)
	protected.Get("/me", h.Me)
	protected.Patch("/me", h.UpdateMe)
}

// @Summary     Sign up
// @Description Creates a new user account with email and password.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RegisterInput true "Registration payload"
// @Success     201  {object} User
// @Failure     400  {object} errs.Error
// @Failure     409  {object} errs.Error "Email already in use"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/signup [post]
func (h *Handler) SignUp(c fiber.Ctx) error {
	var input RegisterInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user, sess, err := h.svc.Register(c.Context(), input, sessionMeta(c))
	if err != nil {
		return h.mapError(err)
	}

	h.setSessionCookie(c, sess)
	return c.Status(fiber.StatusCreated).JSON(user)
}

// @Summary     Log in
// @Description Authenticates with email and password, returns user and sets session cookie.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     LoginInput true "Login payload"
// @Success     200  {object} User
// @Failure     400  {object} errs.Error
// @Failure     401  {object} errs.Error "Invalid credentials"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/login [post]
func (h *Handler) Login(c fiber.Ctx) error {
	var input LoginInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user, sess, err := h.svc.Login(c.Context(), input, sessionMeta(c))
	if err != nil {
		return h.mapError(err)
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
	session := FromContext(c).Session
	if session == nil {
		return errs.BadRequest("no active session")
	}
	if err := h.svc.Logout(c.Context(), session.ID); err != nil {
		return h.mapError(err)
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
	return c.JSON(FromContext(c).User)
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

	user, err := h.svc.UpdateProfile(c.Context(), FromContext(c).User.ID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(user)
}

// @Summary     Verify email
// @Description Confirms a user's email address with a one-time token.
// @Tags        auth
// @Accept      json
// @Param       body body VerifyEmailInput true "Verification token"
// @Success     204  "No Content"
// @Failure     400  {object} errs.Error "Token invalid or expired"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/verify-email [post]
func (h *Handler) VerifyEmail(c fiber.Ctx) error {
	var input VerifyEmailInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.VerifyEmail(c.Context(), input); err != nil {
		return h.mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Resend verification email
// @Description Sends a new email verification token to the given address.
// @Tags        auth
// @Accept      json
// @Param       body body ResendVerificationInput true "Email address"
// @Success     204  "No Content"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     429  {object} errs.Error "Rate limited"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/resend-verification [post]
func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var input ResendVerificationInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.ResendVerification(c.Context(), input); err != nil {
		return h.mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Forgot password
// @Description Sends a password reset email. Always returns 204 to prevent email enumeration.
// @Tags        auth
// @Accept      json
// @Param       body body ForgotPasswordInput true "Email address"
// @Success     204  "No Content"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c fiber.Ctx) error {
	var input ForgotPasswordInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	// Always 204 — never reveal whether the email is registered.
	_ = h.svc.ForgotPassword(c.Context(), input)
	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary     Reset password
// @Description Sets a new password using a valid reset token.
// @Tags        auth
// @Accept      json
// @Param       body body ResetPasswordInput true "Token and new password"
// @Success     204  "No Content"
// @Failure     400  {object} errs.Error "Token invalid or expired"
// @Failure     422  {object} errs.Error "Validation failed"
// @Failure     500  {object} errs.Error
// @Router      /api/v1/auth/reset-password [post]
func (h *Handler) ResetPassword(c fiber.Ctx) error {
	var input ResetPasswordInput
	if err := c.Bind().JSON(&input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	if err := h.svc.ResetPassword(c.Context(), input); err != nil {
		return h.mapError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
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
		return h.mapError(err)
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

	_, sess, err := h.svc.OAuthCallback(c.Context(), provider, code, sessionMeta(c))
	if err != nil {
		return h.mapError(err)
	}

	h.setSessionCookie(c, sess)

	// Redirect to frontend callback — frontend decides routing (onboarding, dashboard, etc.).
	return c.Redirect().To(h.svc.cfg.WebURL + "/auth/callback")
}

// mapError maps domain errors to HTTP errors.
func (h *Handler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrEmailTaken):
		return errs.Conflict("email already in use")
	case errors.Is(err, ErrInvalidCredential):
		return errs.Unauthorized("invalid email or password")
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
func sessionMeta(c fiber.Ctx) SessionMeta {
	return SessionMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	}
}
