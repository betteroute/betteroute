// Package health provides liveness and readiness probes for orchestrators
// such as Kubernetes. Liveness tells the orchestrator whether to restart the
// container; readiness tells it whether to route traffic to this instance.
package health

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// Pinger checks the health of a dependency (e.g. database).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler serves health-check endpoints.
type Handler struct {
	version string
	db      Pinger
}

// New creates a Handler that reports the given build version.
func New(version string, db Pinger) *Handler {
	return &Handler{version: version, db: db}
}

// Register mounts health-check routes on the given app.
func (h *Handler) Register(app *fiber.App) {
	app.Get("/healthz", h.liveness)
	app.Get("/readyz", h.readiness)
}

// @Summary     Liveness probe
// @Description Returns OK if the server is running.
// @Tags        health
// @Produce     json
// @Success     200 {object} object "Status OK with version"
// @Router      /healthz [get]
func (h *Handler) liveness(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "version": h.version})
}

// @Summary     Readiness probe
// @Description Returns OK if the server can serve traffic (database reachable).
// @Tags        health
// @Produce     json
// @Success     200 {object} object "Status OK"
// @Failure     503 {object} object "Database unreachable"
// @Router      /readyz [get]
func (h *Handler) readiness(c fiber.Ctx) error {
	if err := h.db.Ping(c.Context()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unavailable",
			"error":  "database unreachable",
		})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
