// Package migration applies database migrations on startup.
package migration

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	godb "github.com/pressly/goose/v3/database"
	dbpkg "github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationsDir returns the filesystem path to the embedded migrations.
func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	return filepath.Join(dir, "migrations")
}

// Run applies all pending migrations against the database.
// Safe to call multiple times — only unapplied migrations are run.
func Run(ctx context.Context, pool *dbpkg.Pool, logger *slog.Logger) error {
	sqlDB := stdlib.OpenDBFromPool(pool.Pool)
	defer sqlDB.Close()

	store, err := godb.NewStore(godb.DialectPostgres, goose.TableName())
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}

	migrationsDir := migrationsDir()

	// Verify migration files exist
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	hasSQL := false
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			hasSQL = true
			break
		}
	}
	if !hasSQL {
		return fmt.Errorf("no migration .sql files found in %s", migrationsDir)
	}

	p, err := goose.NewProvider(
		"",
		sqlDB,
		os.DirFS(migrationsDir),
		goose.WithStore(store),
	)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	defer p.Close()

	current, target, err := p.GetVersions(ctx)
	if err != nil {
		return fmt.Errorf("get versions: %w", err)
	}

	if current == target && target > 0 {
		logger.Info("database is up to date", "version", current)
		return nil
	}

	logger.Info("running database migrations", "from", current, "to", target)

	results, err := p.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	for _, r := range results {
		logger.Info("migration applied", "version", r.Source.Version, "direction", r.Direction)
	}

	logger.Info("database migrations applied", "version", target)
	return nil
}
