package items

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeStorageLocation_Simple(t *testing.T) {
	require.Equal(t, "drawer", normalizeStorageLocation("drawer"))
	require.Equal(t, "closet", normalizeStorageLocation("  closet  "))
}

func TestNormalizeStorageLocation_MultipleSpaces(t *testing.T) {
	require.Equal(t, "top shelf", normalizeStorageLocation("  top    shelf  "))
	require.Equal(t, "one word", normalizeStorageLocation("one     word"))
}

func TestNormalizeStorageLocation_Empty(t *testing.T) {
	require.Equal(t, "", normalizeStorageLocation(""))
	require.Equal(t, "", normalizeStorageLocation("   "))
	require.Equal(t, "", normalizeStorageLocation("\t\n"))
}

func TestNormalizeStorageLocation_NewlinesAndTabs(t *testing.T) {
	// normalizeStorageLocation only collapses multiple spaces, doesn't convert tabs
	require.Equal(t, "a\tb", normalizeStorageLocation("a\tb"))
	require.Equal(t, "a \t b", normalizeStorageLocation("a \t b"))
}

func TestNormalizeStorageLocation_MixedWhitespace(t *testing.T) {
	require.Equal(t, "hello world", normalizeStorageLocation("  hello   world  "))
}

func TestAppError_Error(t *testing.T) {
	err := ValidationError("test error")
	require.Equal(t, "validation_failed: test error", err.Error())
	require.Equal(t, "validation_failed", err.Code)
	require.Equal(t, "test error", err.Message)
	require.Equal(t, 422, err.Status)
}

func TestAppError_ConflictError(t *testing.T) {
	err := ConflictError("conflict")
	require.Equal(t, "conflict: conflict", err.Error())
	require.Equal(t, "conflict", err.Code)
	require.Equal(t, 409, err.Status)
}

func TestAppError_Details(t *testing.T) {
	err := ValidationError("test")
	err.Details = map[string]interface{}{"field": "name"}
	require.Equal(t, "name", err.Details["field"])
}
