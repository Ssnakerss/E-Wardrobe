package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor is the interface that wraps basic SQL execution methods.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row
}

// Tx represents a database transaction.
type Tx interface {
	Executor
	Begin(ctx context.Context) (pgx.Tx, error)
}

// TxFunc is a function that executes within a transaction.
// The repository receives a transaction-bound query executor.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithTransaction begins a transaction, executes fn, and commits or rolls back.
// The transaction is rolled back on any error or panic.
func WithTransaction(ctx context.Context, pool *Pool, fn TxFunc) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
