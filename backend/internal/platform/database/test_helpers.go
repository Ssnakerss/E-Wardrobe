package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB provides a test PostgreSQL instance via testcontainers.
type TestDB struct {
	Container *postgres.PostgresContainer
	Pool      *Pool
	DSN       string
}

// SetupTestDB creates a new test database with migrations applied.
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	require.NoError(t, err)

	require.NoError(t, pool.Ping(ctx))

	// Apply migrations
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "../../migrations"))

	tDB := &TestDB{
		Container: container,
		Pool:      &Pool{pool},
		DSN:       dsn,
	}

	t.Cleanup(func() {
		tDB.Pool.Close()
	})

	return tDB
}

// ResetDB drops and recreates all tables for a clean state.
func (tDB *TestDB) ResetDB(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	_, err := tDB.Pool.Exec(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
	`)
	require.NoError(t, err)

	db, err := sql.Open("pgx", tDB.DSN)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, goose.Up(db, "../../migrations"))
}
