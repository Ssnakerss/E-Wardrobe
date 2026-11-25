package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// LocalPhotoStorage implements PhotoStorage using the local filesystem.
type LocalPhotoStorage struct {
	baseDir string
	mu      sync.Mutex
}

// NewLocalPhotoStorage creates a new local photo storage implementation.
func NewLocalPhotoStorage(baseDir string) (*LocalPhotoStorage, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("base directory is required")
	}

	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	return &LocalPhotoStorage{
		baseDir: baseDir,
	}, nil
}

// Save writes the content to storage at the given key.
func (s *LocalPhotoStorage) Save(ctx context.Context, key string, content io.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath := s.resolveKey(key)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create temp file first for atomic write
	tmpFile, err := os.CreateTemp(dir, "upload_*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Copy content to temp file
	_, err = io.Copy(tmpFile, content)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write content: %w", err)
	}

	// Close temp file before rename
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}

	// Atomically rename to final path
	if err := os.Rename(tmpName, fullPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename to final path: %w", err)
	}

	return nil
}

// Open reads the content stored at the given key.
func (s *LocalPhotoStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath := s.resolveKey(key)

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

// Delete removes the content at the given key.
func (s *LocalPhotoStorage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath := s.resolveKey(key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// Exists returns true if content exists at the given key.
func (s *LocalPhotoStorage) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath := s.resolveKey(key)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("check existence: %w", err)
}

// resolveKey converts a relative key to an absolute filesystem path.
func (s *LocalPhotoStorage) resolveKey(key string) string {
	// Clean the key to prevent path traversal
	cleanKey := filepath.Clean(key)

	// Ensure the key is relative (no path traversal)
	if filepath.IsAbs(cleanKey) {
		cleanKey = filepath.Base(cleanKey)
	}

	return filepath.Join(s.baseDir, cleanKey)
}

// BaseDir returns the base directory path.
func (s *LocalPhotoStorage) BaseDir() string {
	return s.baseDir
}
