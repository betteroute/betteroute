// Package health provides liveness and readiness probes for orchestrators
// such as Kubernetes. Liveness tells the orchestrator whether to restart the
// container; readiness tells it whether to route traffic to this instance.
package health

import (
	"github.com/gofiber/fiber/v3"
)

// Handler serves health-check endpoints.
type Handler struct {
	version string
}

// New creates a Handler that reports the given build version.
func New(version string) *Handler {
	return &Handler{version: version}
}

// Register mounts health-check routes on the given app.
func (h *Handler) Register(app *fiber.App) {
	app.Get("/healthz", h.liveness)
	app.Get("/readyz", h.readiness)
}

func (h *Handler) liveness(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "version": h.version})
}

func (h *Handler) readiness(c fiber.Ctx) error {
	// TODO: check DB, Redis, etc. once wired.
	return c.JSON(fiber.Map{"status": "ok"})
}
