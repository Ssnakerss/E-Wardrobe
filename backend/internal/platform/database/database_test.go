package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestWithTransaction_Success(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	var executed bool
	err := WithTransaction(ctx, tDB.Pool, func(ctx context.Context, tx pgx.Tx) error {
		executed = true
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})

	require.NoError(t, err)
	require.True(t, executed)
}

func TestWithTransaction_RollbackOnError(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	err := WithTransaction(ctx, tDB.Pool, func(ctx context.Context, tx pgx.Tx) error {
		// This will fail because 'nonexistent_table' doesn't exist
		_, err := tx.Exec(ctx, "SELECT * FROM nonexistent_table_xyz")
		return err
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction failed")
}

func TestWithTransaction_Commit(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	// Insert a value in transaction
	err := WithTransaction(ctx, tDB.Pool, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
			uuid.New(), "test@test.com", "hash")
		return err
	})
	require.NoError(t, err)

	// Verify it was committed
	var count int
	err = tDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", "test@test.com").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestWithTransaction_PanicRecovers(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	require.Panics(t, func() {
		_ = WithTransaction(ctx, tDB.Pool, func(ctx context.Context, tx pgx.Tx) error {
			panic("test panic")
		})
	})
}

func TestPool_Ready(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, tDB.Pool.Ready(ctx))
}

func TestPool_QueryRow(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	var result int
	err := tDB.Pool.QueryRowContext(ctx, "SELECT 42").Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 42, result)
}

func TestPool_Exec(t *testing.T) {
	tDB := SetupTestDB(t)
	ctx := context.Background()

	tag, err := tDB.Pool.ExecContext(ctx, "SELECT 1")
	require.NoError(t, err)
	require.NotNil(t, tag)
}

