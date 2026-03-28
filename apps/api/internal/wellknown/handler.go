package wellknown

import (
	"github.com/gofiber/fiber/v3"

	"github.com/execrc/betteroute/internal/errs"
)

// Handler serves .well-known files for iOS Universal Links and Android App Links.
type Handler struct {
	store *Store
}

// NewHandler creates a new well-known handler.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// Register mounts /.well-known routes.
// Must be registered BEFORE the catch-all redirect route.
func (h *Handler) Register(r fiber.Router) {
	wk := r.Group("/.well-known")
	wk.Get("/apple-app-site-association", h.AASA)
	wk.Get("/assetlinks.json", h.AssetLinks)
}

// @Summary     Apple App Site Association
// @Description Serves the AASA file for iOS Universal Links based on the request hostname.
// @Tags        well-known
// @Produce     json
// @Success     200 {object} aasaResponse
// @Failure     500 {object} errs.Error
// @Router      /.well-known/apple-app-site-association [get]
func (h *Handler) AASA(c fiber.Ctx) error {
	apps, err := h.store.FindWorkspaceApps(c.Context(), c.Hostname(), "ios")
	if err != nil {
		return errs.Internal("").WithCause(err)
	}

	details := make([]aasaDetail, 0, len(apps))
	for _, app := range apps {
		if app.TeamID == nil || app.BundleID == nil {
			continue
		}
		details = append(details, aasaDetail{
			AppIDs:     []string{*app.TeamID + "." + *app.BundleID},
			Components: []aasaComponent{{Path: "/*"}},
		})
	}

	c.Set("Content-Type", "application/json")
	c.Set("Cache-Control", "public, max-age=3600")
	return c.JSON(aasaResponse{AppLinks: aasaAppLinks{Details: details}})
}

// @Summary     Android Asset Links
// @Description Serves the assetlinks.json file for Android App Links based on the request hostname.
// @Tags        well-known
// @Produce     json
// @Success     200 {array}  assetLinkStatement
// @Failure     500 {object} errs.Error
// @Router      /.well-known/assetlinks.json [get]
func (h *Handler) AssetLinks(c fiber.Ctx) error {
	apps, err := h.store.FindWorkspaceApps(c.Context(), c.Hostname(), "android")
	if err != nil {
		return errs.Internal("").WithCause(err)
	}

	statements := make([]assetLinkStatement, 0, len(apps))
	for _, app := range apps {
		if app.PackageName == nil {
			continue
		}
		fingerprints := app.SHA256Fingerprints
		if fingerprints == nil {
			fingerprints = []string{}
		}
		statements = append(statements, assetLinkStatement{
			Relation: []string{"delegate_permission/common.handle_all_urls"},
			Target: assetLinkTarget{
				Namespace:              "android_app",
				PackageName:            *app.PackageName,
				SHA256CertFingerprints: fingerprints,
			},
		})
	}

	c.Set("Content-Type", "application/json")
	c.Set("Cache-Control", "public, max-age=3600")
	return c.JSON(statements)
}
