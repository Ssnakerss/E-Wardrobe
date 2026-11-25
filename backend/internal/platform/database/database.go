package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with additional lifecycle methods.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new PostgreSQL connection pool from the given DSN.
func NewPool(ctx context.Context, dsn string, log *slog.Logger) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("connected to database", "dsn", dsn)

	return &Pool{pool}, nil
}

// Ready checks if the database is reachable and responsive.
func (p *Pool) Ready(ctx context.Context) error {
	return p.Ping(ctx)
}

// ExecContext executes a raw SQL query.
func (p *Pool) ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, query, args...)
}

// QueryContext queries and returns rows.
func (p *Pool) QueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return p.Pool.Query(ctx, query, args...)
}

// QueryRowContext queries and returns a single row.
func (p *Pool) QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
	return p.Pool.QueryRow(ctx, query, args...)
}
