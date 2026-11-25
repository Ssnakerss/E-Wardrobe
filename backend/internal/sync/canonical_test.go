package sync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidEntityType(t *testing.T) {
	require.True(t, ValidEntityType("category"))
	require.True(t, ValidEntityType("item"))
	require.True(t, ValidEntityType("photo"))
	require.False(t, ValidEntityType("user"))
	require.False(t, ValidEntityType(""))
	require.False(t, ValidEntityType("unknown"))
}

func TestValidOperation(t *testing.T) {
	require.True(t, ValidOperation("create"))
	require.True(t, ValidOperation("update"))
	require.True(t, ValidOperation("delete"))
	require.True(t, ValidOperation("archive"))
	require.True(t, ValidOperation("unarchive"))
	require.False(t, ValidOperation("insert"))
	require.False(t, ValidOperation("patch"))
	require.False(t, ValidOperation(""))
}

func TestCanonicalHash_Deterministic(t *testing.T) {
	payload := map[string]any{
		"name":    "test",
		"version": int64(1),
		"active":  true,
	}
	h1 := CanonicalHash(payload)
	h2 := CanonicalHash(payload)
	require.Equal(t, h1, h2)
	require.NotEmpty(t, h1)
	require.Len(t, h1, 64) // SHA-256 hex
}

func TestCanonicalHash_SameContentDifferentOrder(t *testing.T) {
	// Keys in different order should produce same hash
	p1 := map[string]any{
		"b": "2",
		"a": "1",
		"c": "3",
	}
	p2 := map[string]any{
		"c": "3",
		"a": "1",
		"b": "2",
	}
	require.Equal(t, CanonicalHash(p1), CanonicalHash(p2))
}

func TestCanonicalHash_DifferentContent(t *testing.T) {
	p1 := map[string]any{"name": "a"}
	p2 := map[string]any{"name": "b"}
	require.NotEqual(t, CanonicalHash(p1), CanonicalHash(p2))
}

func TestCanonicalHash_NestedMaps(t *testing.T) {
	p1 := map[string]any{
		"outer": map[string]any{
			"z": 2,
			"a": 1,
		},
	}
	p2 := map[string]any{
		"outer": map[string]any{
			"a": 1,
			"z": 2,
		},
	}
	require.Equal(t, CanonicalHash(p1), CanonicalHash(p2))
}

func TestCanonicalHash_EmptyPayload(t *testing.T) {
	h := CanonicalHash(map[string]any{})
	require.NotEmpty(t, h)
	require.Len(t, h, 64)
}

func TestCanonicalHash_NullValues(t *testing.T) {
	p := map[string]any{
		"name": "test",
		"null": nil,
	}
	h := CanonicalHash(p)
	require.NotEmpty(t, h)
}

func TestCanonicalHash_ArrayValues(t *testing.T) {
	p := map[string]any{
		"tags": []any{"a", "b", "c"},
	}
	h1 := CanonicalHash(p)
	h2 := CanonicalHash(p)
	require.Equal(t, h1, h2)
}

func TestCanonicalHash_NumberTypes(t *testing.T) {
	// int64 and float64 representations should be normalized consistently
	p1 := map[string]any{"count": int64(42)}
	p2 := map[string]any{"count": float64(42)}
	// After normalization, both should produce the same hash
	require.Equal(t, CanonicalHash(p1), CanonicalHash(p2))
}

func TestCanonicalHash_LargePayload(t *testing.T) {
	p := make(map[string]any)
	for i := 0; i < 100; i++ {
		p[string(rune('a'+i%26))] = i
	}
	h := CanonicalHash(p)
	require.NotEmpty(t, h)
	require.Len(t, h, 64)
}

func TestToSyncOperation(t *testing.T) {
	require.Equal(t, "upsert", toSyncOperation("create"))
	require.Equal(t, "upsert", toSyncOperation("update"))
	require.Equal(t, "delete", toSyncOperation("delete"))
	require.Equal(t, "delete", toSyncOperation("archive"))
	require.Equal(t, "delete", toSyncOperation("unarchive"))
	require.Equal(t, "upsert", toSyncOperation("unknown"))
}

func TestContainsStr(t *testing.T) {
	require.True(t, containsStr("hello world", "world"))
	require.True(t, containsStr("hello world", "hello"))
	require.True(t, containsStr("hello world", "hello world"))
	require.True(t, containsStr("hello world", ""))
	require.False(t, containsStr("hello world", "bye"))
	require.False(t, containsStr("hello", "world"))
	require.False(t, containsStr("", "a"))
}

func TestLowercase(t *testing.T) {
	require.Equal(t, "hello world", lowercase("HELLO WORLD"))
	require.Equal(t, "hello world", lowercase("Hello World"))
	require.Equal(t, "", lowercase(""))
	require.Equal(t, "abc123", lowercase("ABC123"))
}
