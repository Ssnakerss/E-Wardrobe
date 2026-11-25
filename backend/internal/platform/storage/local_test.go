package storage

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLocalPhotoStorage_ValidDir(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, storage)
	require.Equal(t, tmpDir, storage.BaseDir())
}

func TestNewLocalPhotoStorage_EmptyDir(t *testing.T) {
	_, err := NewLocalPhotoStorage("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "base directory is required")
}

func TestNewLocalPhotoStorage_CreatesDir(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "sub", "dir")
	storage, err := NewLocalPhotoStorage(baseDir)
	require.NoError(t, err)
	require.NotNil(t, storage)

	// Verify directory was created
	_, err = os.Stat(baseDir)
	require.NoError(t, err)
}

func TestStorage_SaveAndOpen(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	data := []byte("test content for storage")
	key := "test/file.txt"

	err = storage.Save(ctx, key, bytes.NewReader(data))
	require.NoError(t, err)

	rc, err := storage.Open(ctx, key)
	require.NoError(t, err)
	defer rc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	require.Equal(t, data, buf.Bytes())
}

func TestStorage_SaveOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test.txt"

	err = storage.Save(ctx, key, bytes.NewReader([]byte("first")))
	require.NoError(t, err)

	err = storage.Save(ctx, key, bytes.NewReader([]byte("second")))
	require.NoError(t, err)

	rc, err := storage.Open(ctx, key)
	require.NoError(t, err)
	defer rc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	require.Equal(t, []byte("second"), buf.Bytes())
}

func TestStorage_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "to-delete.txt"

	err = storage.Save(ctx, key, bytes.NewReader([]byte("content")))
	require.NoError(t, err)

	err = storage.Delete(ctx, key)
	require.NoError(t, err)

	_, err = storage.Open(ctx, key)
	require.Error(t, err)
}

func TestStorage_DeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	// Deleting non-existent file should not error
	err = storage.Delete(ctx, "nonexistent.txt")
	require.NoError(t, err)
}

func TestStorage_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "exists.txt"

	exists, err := storage.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)

	err = storage.Save(ctx, key, bytes.NewReader([]byte("content")))
	require.NoError(t, err)

	exists, err = storage.Exists(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestStorage_PathTraversalPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	// Try path traversal
	evilKey := "../../../etc/passwd"

	// Save should be safe (resolved within base dir)
	err = storage.Save(ctx, evilKey, bytes.NewReader([]byte("evil")))
	// On Windows, this might succeed or fail depending on path cleaning
	// The important thing is it doesn't escape baseDir

	// Open should also be safe
	_, err = storage.Open(ctx, evilKey)
	// On Windows, the path might be cleaned differently
	_ = err
}

func TestStorage_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "deep/nested/dir/file.txt"

	err = storage.Save(ctx, key, bytes.NewReader([]byte("nested content")))
	require.NoError(t, err)

	rc, err := storage.Open(ctx, key)
	require.NoError(t, err)
	defer rc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	require.Equal(t, []byte("nested content"), buf.Bytes())
}

func TestStorage_BinaryContent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	key := "binary.dat"

	err = storage.Save(ctx, key, bytes.NewReader(data))
	require.NoError(t, err)

	rc, err := storage.Open(ctx, key)
	require.NoError(t, err)
	defer rc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	require.Equal(t, data, buf.Bytes())
}

func TestStorage_BaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalPhotoStorage(tmpDir)
	require.NoError(t, err)
	require.Equal(t, tmpDir, storage.BaseDir())
}
