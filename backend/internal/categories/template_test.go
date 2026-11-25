package categories

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrapTemplate_NonEmpty(t *testing.T) {
	require.Greater(t, len(BootstrapTemplate), 0)
}

func TestBootstrapTemplate_NoDuplicates(t *testing.T) {
	names := make(map[string]bool)
	for _, cat := range BootstrapTemplate {
		require.False(t, names[cat.Name], "duplicate category name: %s", cat.Name)
		names[cat.Name] = true
	}
}

func TestBootstrapTemplate_AllHaveNames(t *testing.T) {
	for _, cat := range BootstrapTemplate {
		require.NotEmpty(t, cat.Name)
	}
}

func TestBootstrapTemplate_SortOrder(t *testing.T) {
	for i, cat := range BootstrapTemplate {
		require.Equal(t, int32(i+1), cat.SortOrder)
	}
}

func TestBootstrapTemplate_NoParentIDs(t *testing.T) {
	for _, cat := range BootstrapTemplate {
		require.Nil(t, cat.ParentID)
	}
}
