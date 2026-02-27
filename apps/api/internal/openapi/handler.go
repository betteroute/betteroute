package openapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/swaggo/swag"
)

// Register mounts the API documentation routes.
// GET /docs       → Scalar UI (interactive API explorer)
// GET /docs/json  → raw OpenAPI spec
func Register(app *fiber.App) {
	spec, err := swag.ReadDoc()
	if err != nil {
		spec = `{"error":"failed to read OpenAPI spec"}`
	}

	app.Get("/docs/json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(spec)
	})

	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(scalarHTML)
	})
}

const scalarHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Betteroute API</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script id="api-reference" data-url="/docs/json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
