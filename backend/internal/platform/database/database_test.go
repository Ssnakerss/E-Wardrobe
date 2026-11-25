package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/stretchr/testify/require"
)

func TestPoolReady(t *testing.T) {
	tDB := database.SetupTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, tDB.Pool.Ready(ctx))
}

func TestPoolPing(t *testing.T) {
	tDB := database.SetupTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, tDB.Pool.Ping(ctx))
}
