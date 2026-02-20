package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/errs"
)

const sessionCookie = "session"

// Auth validates the session cookie and loads the authenticated user and
// session into request locals.
func Auth(svc *auth.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(sessionCookie)
		if token == "" {
			return errs.Unauthorized("")
		}

		user, sess, err := svc.ValidateSession(c.Context(), token)
		if err != nil {
			return errs.Unauthorized("")
		}

		c.Locals("user", user)
		c.Locals("session", sess)

		return c.Next()
	}
}
