package changejournal

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChange_Structure(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	change := Change{
		UserID:        id,
		Seq:           42,
		Entity:        "item",
		EntityID:      id,
		Operation:     "upsert",
		EntityVersion: 1,
		Payload:       map[string]any{"name": "test"},
		ChangedAt:     now,
	}

	require.Equal(t, id, change.UserID)
	require.Equal(t, int64(42), change.Seq)
	require.Equal(t, "item", change.Entity)
	require.Equal(t, "upsert", change.Operation)
	require.Equal(t, int64(1), change.EntityVersion)
	require.Equal(t, now, change.ChangedAt)
}

func TestChange_EmptyPayload(t *testing.T) {
	change := Change{
		Entity:        "category",
		EntityID:      uuid.New(),
		Operation:     "delete",
		EntityVersion: 1,
		Payload:       nil,
	}
	require.Nil(t, change.Payload)
	require.Equal(t, "delete", change.Operation)
}

func TestChange_ZeroValue(t *testing.T) {
	var change Change
	require.Equal(t, uuid.Nil, change.UserID)
	require.Equal(t, int64(0), change.Seq)
	require.Equal(t, "", change.Entity)
	require.Equal(t, "", change.Operation)
	require.Equal(t, int64(0), change.EntityVersion)
	require.Equal(t, time.Time{}, change.ChangedAt)
}
