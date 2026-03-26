// Package middleware provides HTTP middleware for authentication, logging, and other cross-cutting concerns.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/apikey"
	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/guard"
)

// Auth validates authentication via session cookie or API key bearer token.
// Session auth injects auth.Context{User, Session}. API key auth injects both
// auth.Context{User} (key creator) and apikey.Context (the key itself).
func Auth(authSvc *auth.Service, apikeySvc *apikey.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Try Bearer token first (API key) — single JOIN query for key + creator.
		if bearer := parseBearer(c); bearer != "" {
			return authViaAPIKey(c, apikeySvc, bearer)
		}

		// Fall back to session cookie.
		return authViaSession(c, authSvc)
	}
}

// authViaSession validates the session cookie and injects auth.Context.
func authViaSession(c fiber.Ctx, svc *auth.Service) error {
	token := c.Cookies(auth.CookieName)
	if token == "" {
		return errs.Unauthorized("")
	}

	user, sess, err := svc.ValidateSession(c.Context(), token)
	if err != nil {
		return errs.Unauthorized("")
	}

	c.SetContext(auth.NewContext(c.Context(), auth.Context{
		User:    user,
		Session: sess,
	}))

	return c.Next()
}

// authViaAPIKey validates a Bearer token as an API key and injects
// both auth.Context (key creator) and apikey.Context (key itself).
// Uses a single JOIN query to resolve key + creator in one round-trip.
func authViaAPIKey(c fiber.Ctx, apikeySvc *apikey.Service, plain string) error {
	key, creator, err := apikeySvc.ValidateKeyWithCreator(c.Context(), plain)
	if err != nil {
		return errs.Unauthorized("")
	}

	if creator.Status != "active" {
		return errs.Unauthorized("")
	}

	user := &auth.User{
		ID:              creator.ID,
		Name:            creator.Name,
		Email:           creator.Email,
		EmailVerifiedAt: creator.EmailVerifiedAt,
		AvatarURL:       creator.AvatarURL,
		Status:          creator.Status,
		OnboardedAt:     creator.OnboardedAt,
		Timezone:        creator.Timezone,
		LastLoginAt:     creator.LastLoginAt,
		CreatedAt:       creator.CreatedAt,
		UpdatedAt:       creator.UpdatedAt,
	}

	// Inject auth context (key acts on behalf of creator).
	c.SetContext(auth.NewContext(c.Context(), auth.Context{
		User: user,
	}))

	// Inject API key context — downstream can check apikey.FromContext(ctx) != nil.
	c.SetContext(apikey.NewContext(c.Context(), key))

	// Inject scope checker so guard.Scope() can verify API key permissions.
	c.SetContext(guard.WithScope(c.Context(), key))

	return c.Next()
}

// parseBearer returns the bearer token if it starts with the API key prefix.
func parseBearer(c fiber.Ctx) string {
	h := c.Get("Authorization")
	if h == "" {
		return ""
	}
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(h, bearerPrefix) {
		return ""
	}
	token := h[len(bearerPrefix):]
	if !strings.HasPrefix(token, apikey.Prefix) {
		return ""
	}
	return token
}
