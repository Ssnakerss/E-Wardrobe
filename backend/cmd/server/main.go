package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/categories"
	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/config"
	"github.com/Ssnakerss/E-Wardrobe/internal/httpapi"
	"github.com/Ssnakerss/E-Wardrobe/internal/items"
	"github.com/Ssnakerss/E-Wardrobe/internal/photos"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/storage"
	"github.com/Ssnakerss/E-Wardrobe/internal/webui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)

	pool, err := database.NewPool(context.Background(), cfg.DatabaseURL, logger)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Initialize storage
	store, err := storage.NewLocalPhotoStorage(cfg.PhotoStorageDir)
	if err != nil {
		slog.Error("failed to create photo storage", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	catRepo := categories.NewRepository(pool)
	itemRepo := items.NewRepository(pool)
	photoRepo := photos.NewRepository(pool)

	// Initialize change journal
	journal := changejournal.NewWriter(pool)

	// Initialize services
	itemSvc := items.NewService(itemRepo, catRepo, journal)
	photoSvc := photos.NewService(photoRepo, store, journal, cfg.PhotoStorageDir)

	// Build router config
	routerCfg := httpapi.RouterConfig{
		Pool:      pool,
		Store:     store,
		TmpDir:    cfg.PhotoStorageDir,
		Journal:   journal,
		ItemSvc:   itemSvc,
		PhotoSvc:  photoSvc,
		PhotoRepo: photoRepo,
	}

	// Main router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health endpoints
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := pool.Ready(ctx); err != nil {
			slog.Error("ready check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	// API router
	apiRouter := httpapi.NewRouter(routerCfg)
	r.Mount("/", apiRouter)

	// Static web files (SPA fallback) — must be after API routes
	r.Handle("/*", webui.StaticHandler())

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting HTTP server", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
