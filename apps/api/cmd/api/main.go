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
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/config"
	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/docs"
	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/folder"
	"github.com/execrc/betteroute/internal/health"
	"github.com/execrc/betteroute/internal/link"
	"github.com/execrc/betteroute/internal/openapi"
	"github.com/execrc/betteroute/internal/redirect"
	"github.com/execrc/betteroute/internal/tag"
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

	registerRoutes(app, cfg, logger, pool)

	return serve(app, cfg, logger)
}

// registerRoutes wires all handler packages to the Fiber app.
func registerRoutes(app *fiber.App, cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) {
	health.New(config.Version, pool).Register(app)

	if cfg.IsDevelopment() {
		docs.Register(app, openapi.Spec)
		logger.Info("API docs available at /docs")
	}

	// Links.
	linkStore := link.NewStore(pool)
	linkSvc := link.NewService(linkStore)
	linkHandler := link.NewHandler(linkSvc)

	// Folders.
	folderStore := folder.NewStore(pool)
	folderSvc := folder.NewService(folderStore)
	folderHandler := folder.NewHandler(folderSvc)

	// Tags.
	tagStore := tag.NewStore(pool)
	tagSvc := tag.NewService(tagStore)
	tagHandler := tag.NewHandler(tagSvc)

	// API v1.
	api := app.Group("/api/v1")
	linkHandler.Register(api)
	folderHandler.Register(api)
	tagHandler.Register(api)

	// Tag-link association routes: /api/v1/links/:id/tags
	tagHandler.RegisterLinkRoutes(api.Group("/links"))

	// Redirect — catch-all, registered last.
	redirectSvc := redirect.NewService(pool)
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
