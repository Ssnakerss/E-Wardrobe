package storage

import (
	"context"
	"io"
)

// PhotoStorage defines the interface for persistent photo content storage.
// Implementations may use the local filesystem, S3, or any other backend.
type PhotoStorage interface {
	// Save writes the content to storage at the given key.
	Save(ctx context.Context, key string, content io.Reader) error

	// Open reads the content stored at the given key.
	// Returns os.ErrNotFound-style errors when the key does not exist.
	Open(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the content at the given key.
	Delete(ctx context.Context, key string) error

	// Exists returns true if content exists at the given key.
	Exists(ctx context.Context, key string) (bool, error)
}
