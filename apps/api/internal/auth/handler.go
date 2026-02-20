package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/validate"
)

const cookieName = "session"

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

	// Protected routes — require a valid session cookie.
	protected := a.Group("", authMW)
	protected.Post("/logout", h.Logout)
	protected.Get("/me", h.Me)
	protected.Patch("/me", h.UpdateMe)
}

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

func (h *Handler) Logout(c fiber.Ctx) error {
	sess := SessionFromContext(c)
	if err := h.svc.Logout(c.Context(), sess.ID); err != nil {
		return h.mapError(err)
	}
	h.clearSessionCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Me(c fiber.Ctx) error {
	// User is already loaded by Authenticate middleware.
	return c.JSON(UserFromContext(c))
}

func (h *Handler) UpdateMe(c fiber.Ctx) error {
	var input UpdateProfileInput
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return errs.BadRequest("invalid request body")
	}
	if fieldErrs := validate.Struct(input); fieldErrs != nil {
		return errs.Validation(fieldErrs)
	}

	user, err := h.svc.UpdateProfile(c.Context(), UserFromContext(c).ID, input)
	if err != nil {
		return h.mapError(err)
	}

	return c.JSON(user)
}

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
		Name:     cookieName,
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
		Name:     cookieName,
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
