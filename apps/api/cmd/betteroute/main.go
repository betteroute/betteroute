// @title Betteroute API
// @version 0.1.0
// @description Open-source link management platform.

// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/apikey"
	"github.com/execrc/betteroute/internal/auth"
	"github.com/execrc/betteroute/internal/config"
	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/folder"
	"github.com/execrc/betteroute/internal/health"
	"github.com/execrc/betteroute/internal/link"
	"github.com/execrc/betteroute/internal/middleware"
	"github.com/execrc/betteroute/internal/notify"
	"github.com/execrc/betteroute/internal/notify/email"
	"github.com/execrc/betteroute/internal/openapi"
	"github.com/execrc/betteroute/internal/redirect"
	"github.com/execrc/betteroute/internal/sqlc"
	"github.com/execrc/betteroute/internal/tag"
	"github.com/execrc/betteroute/internal/workspace"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Logger: text + source in development, JSON in production.
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	var handler slog.Handler
	if cfg.IsDevelopment() {
		opts.Level = slog.LevelDebug
		opts.AddSource = true
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Database.
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	logger.Info("database connected")

	// Fiber app.
	app := fiber.New(fiber.Config{
		AppName:      "betteroute-api",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: errs.Handler(logger),
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.WebURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	registerRoutes(app, cfg, logger, pool)

	return serve(app, cfg, logger)
}

// registerRoutes wires all handler packages to the Fiber app.
func registerRoutes(app *fiber.App, cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) {
	health.New(config.Version, pool).Register(app)

	if cfg.IsDevelopment() {
		openapi.Register(app)
		logger.Info("API docs available at /docs")
	}

	// Notifier.
	var notifier notify.Notifier
	if cfg.EmailAPIKey != "" {
		notifier = email.New(cfg.EmailAPIKey, cfg.EmailFrom)
	} else {
		notifier = notify.Nop()
		logger.Warn("EMAIL_API_KEY not set, notifications will be dropped")
	}

	// Services.
	authSvc := auth.NewService(auth.NewStore(pool), notifier, auth.Config{
		WebURL:             cfg.WebURL,
		APIURL:             cfg.APIURL,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
	})
	wsSvc := workspace.NewService(workspace.NewStore(pool), notifier, cfg.WebURL)
	linkSvc := link.NewService(link.NewStore(pool))
	folderSvc := folder.NewService(folder.NewStore(pool))
	tagSvc := tag.NewService(tag.NewStore(pool))
	apikeySvc := apikey.NewService(apikey.NewStore(pool))
	redirectSvc := redirect.NewService(pool)

	// Middleware.
	authMW := middleware.Auth(authSvc, apikeySvc)
	workspaceMW := middleware.Workspace(wsSvc)
	entitlementMW := middleware.Entitlement(sqlc.New(pool))

	// Handlers.
	authHandler := auth.NewHandler(authSvc, !cfg.IsDevelopment())
	wsHandler := workspace.NewHandler(wsSvc)
	linkHandler := link.NewHandler(linkSvc, tagSvc)
	folderHandler := folder.NewHandler(folderSvc)
	tagHandler := tag.NewHandler(tagSvc)
	apikeyHandler := apikey.NewHandler(apikeySvc)

	// Routes.
	//
	// Middleware chain per workspace-scoped request:
	//   Auth → Workspace → Entitlement → Handler
	//
	// Authorization (role, scope, quota, feature) is checked inside each
	// handler method via the guard package — not via per-route middleware.
	api := app.Group("/api/v1")
	authHandler.Register(api, authMW)

	api.Use(authMW)

	wsHandler.Register(api, workspaceMW, entitlementMW)

	ws := api.Group("/workspaces/:slug", workspaceMW, entitlementMW)

	linkHandler.Register(ws.Group("/links"))
	folderHandler.Register(ws.Group("/folders"))
	tagHandler.Register(ws.Group("/tags"))
	apikeyHandler.Register(ws.Group("/api-keys"))

	// Redirect — catch-all, registered last.
	redirect.NewHandler(redirectSvc).Register(app)

	logger.Debug("routes registered")
}

// serve starts the HTTP listener and blocks until a shutdown signal is received
// or a listen error occurs. It performs graceful shutdown with a 10 s deadline.
func serve(app *fiber.App, cfg *config.Config, logger *slog.Logger) error {
	addr := fmt.Sprintf(":%d", cfg.Port)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", addr, "env", cfg.Env)
		errCh <- app.Listen(addr, fiber.ListenConfig{
			DisableStartupMessage: !cfg.IsDevelopment(),
		})
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server listen: %w", err)
	case <-quit:
		logger.Info("shutting down server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
