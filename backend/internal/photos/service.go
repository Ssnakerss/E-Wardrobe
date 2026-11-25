package photos

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/storage"
	"github.com/google/uuid"
)

// Service handles business logic for photos.
type Service struct {
	repo    *Repository
	store   storage.PhotoStorage
	journal *changejournal.Writer
	tmpDir  string
}

// NewService creates a new photos service.
func NewService(repo *Repository, store storage.PhotoStorage, journal *changejournal.Writer, tmpDir string) *Service {
	return &Service{repo: repo, store: store, journal: journal, tmpDir: tmpDir}
}

// Create creates a new photo metadata record with awaiting_upload status.
func (s *Service) Create(ctx context.Context, tx database.Tx, userID, itemID uuid.UUID, input PhotoInput) (Photo, error) {
	photo := Photo{
		ID:        uuid.New(),
		ItemID:    itemID,
		UserID:    userID,
		SortOrder: input.SortOrder,
		Status:    PhotoStatusAwaitingUpload,
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	photo, err := s.repo.Create(ctx, tx, photo)
	if err != nil {
		return Photo{}, fmt.Errorf("create photo: %w", err)
	}

	// Write to change journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "photo",
		EntityID:      photo.ID,
		Operation:     "upsert",
		EntityVersion: photo.Version,
		Payload:       photoToPayload(photo),
	})
	if err != nil {
		return photo, fmt.Errorf("write journal: %w", err)
	}

	return photo, nil
}

// GetByID returns a photo by ID with ownership check.
func (s *Service) GetByID(ctx context.Context, tx database.Tx, id, userID uuid.UUID) (Photo, error) {
	return s.repo.GetByID(ctx, tx, id, userID)
}

// ListByItem returns all photos for an item.
func (s *Service) ListByItem(ctx context.Context, tx database.Tx, itemID, userID uuid.UUID) ([]Photo, error) {
	return s.repo.GetByItem(ctx, tx, itemID, userID)
}

// Update updates a photo metadata record.
func (s *Service) Update(ctx context.Context, tx database.Tx, userID uuid.UUID, patch PhotoPatch) (Photo, error) {
	// Verify ownership
	existing, err := s.repo.GetByID(ctx, tx, patch.ID, userID)
	if err != nil {
		return Photo{}, fmt.Errorf("photo not found or not owned: %w", err)
	}

	// Check version
	if patch.BaseVersion != 0 && patch.BaseVersion != existing.Version {
		return Photo{}, fmt.Errorf("version conflict: expected %d, got %d", existing.Version, patch.BaseVersion)
	}

	updated, err := s.repo.Update(ctx, tx, patch)
	if err != nil {
		return Photo{}, fmt.Errorf("update photo: %w", err)
	}

	// Write to change journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "photo",
		EntityID:      updated.ID,
		Operation:     "upsert",
		EntityVersion: updated.Version,
		Payload:       photoToPayload(updated),
	})
	if err != nil {
		return updated, fmt.Errorf("write journal: %w", err)
	}

	return updated, nil
}

// SoftDelete marks a photo as deleted.
func (s *Service) SoftDelete(ctx context.Context, tx database.Tx, id, userID uuid.UUID, version int64) error {
	// Verify ownership
	_, err := s.repo.GetByID(ctx, tx, id, userID)
	if err != nil {
		return fmt.Errorf("photo not found or not owned: %w", err)
	}

	err = s.repo.SoftDelete(ctx, tx, id, userID, version)
	if err != nil {
		return fmt.Errorf("soft delete photo: %w", err)
	}

	// Write tombstone to journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "photo",
		EntityID:      id,
		Operation:     "delete",
		EntityVersion: version + 1,
		Payload:       map[string]any{"tombstone": true},
	})
	if err != nil {
		return fmt.Errorf("write tombstone: %w", err)
	}

	return nil
}

// UpdateSortOrder updates the sort order of a photo.
func (s *Service) UpdateSortOrder(ctx context.Context, tx database.Tx, userID uuid.UUID, id uuid.UUID, sortOrder int32, version int64) (Photo, error) {
	// Verify ownership
	_, err := s.repo.GetByID(ctx, tx, id, userID)
	if err != nil {
		return Photo{}, fmt.Errorf("photo not found or not owned: %w", err)
	}

	updated, err := s.repo.UpdateSortOrder(ctx, tx, id, userID, sortOrder, version)
	if err != nil {
		return Photo{}, fmt.Errorf("update photo sort order: %w", err)
	}

	// Write to change journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "photo",
		EntityID:      updated.ID,
		Operation:     "upsert",
		EntityVersion: updated.Version,
		Payload:       photoToPayload(updated),
	})
	if err != nil {
		return updated, fmt.Errorf("write journal: %w", err)
	}

	return updated, nil
}

// UpdatePhotoOrderBatch atomically updates sort orders for multiple photos.
func (s *Service) UpdatePhotoOrderBatch(ctx context.Context, tx database.Tx, userID, itemID uuid.UUID, orders []PhotoOrder) error {
	// Verify all photos belong to the user and item
	for _, order := range orders {
		photo, err := s.repo.GetByID(ctx, tx, order.ID, userID)
		if err != nil {
			return fmt.Errorf("photo %s not found or not owned: %w", order.ID, err)
		}
		if photo.ItemID != itemID {
			return fmt.Errorf("photo %s does not belong to item %s", order.ID, itemID)
		}
	}

	// Update each photo's sort order
	for _, order := range orders {
		_, err := s.repo.UpdateSortOrder(ctx, tx, order.ID, userID, order.SortOrder, 0)
		if err != nil {
			return fmt.Errorf("update photo %s sort order: %w", order.ID, err)
		}
	}

	return nil
}

// GetContent returns a reader for the photo content.
func (s *Service) GetContent(ctx context.Context, tx database.Tx, id, userID uuid.UUID, variant string) (io.ReadCloser, error) {
	// Get photo to verify ownership and get keys
	photo, err := s.repo.GetByID(ctx, tx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("photo not found or not owned: %w", err)
	}

	// Determine which key to use
	var key string
	switch variant {
	case "thumbnail":
		if photo.ThumbnailKey == nil {
			return nil, fmt.Errorf("thumbnail not available")
		}
		key = *photo.ThumbnailKey
	case "original", "":
		if photo.ObjectKey == nil {
			return nil, fmt.Errorf("original not available")
		}
		key = *photo.ObjectKey
	default:
		return nil, fmt.Errorf("invalid variant: %s", variant)
	}

	rc, err := s.store.Open(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("open photo content: %w", err)
	}

	return rc, nil
}

// TemporaryFile creates a temporary file for the upload pipeline.
func (s *Service) TemporaryFile() (*os.File, error) {
	if s.tmpDir == "" {
		var err error
		s.tmpDir, err = os.MkdirTemp("", "wardrobe_photos_")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
	}

	f, err := os.CreateTemp(s.tmpDir, "upload_*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	return f, nil
}

// CleanupTemp removes temporary files.
func (s *Service) CleanupTemp() error {
	if s.tmpDir != "" {
		return os.RemoveAll(s.tmpDir)
	}
	return nil
}

// GetTempDir returns the temp directory path.
func (s *Service) GetTempDir() string {
	return s.tmpDir
}

// EnsureTempDir creates the temp directory if it doesn't exist.
func (s *Service) EnsureTempDir() error {
	if s.tmpDir == "" {
		dir, err := os.MkdirTemp("", "wardrobe_photos_")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		s.tmpDir = dir
	}

	if _, err := os.Stat(s.tmpDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.tmpDir, 0755); err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
	}

	return nil
}

// photoToPayload converts a photo to a journal payload map.
func photoToPayload(p Photo) map[string]any {
	return map[string]any{
		"entity_id":    p.ID,
		"item_id":      p.ItemID,
		"sort_order":   p.SortOrder,
		"status":       string(p.Status),
		"mime_type":    p.MimeType,
		"size_bytes":   p.SizeBytes,
		"width":        p.Width,
		"height":       p.Height,
		"content_hash": p.ContentHash,
		"version":      p.Version,
	}
}

// ObjectKey returns the storage key for the original photo.
func ObjectKey(userID, itemID, photoID uuid.UUID) string {
	return filepath.Join("users", userID.String(), "items", itemID.String(), photoID.String(), "original.jpg")
}

// ThumbnailKey returns the storage key for the thumbnail.
func ThumbnailKey(userID, itemID, photoID uuid.UUID) string {
	return filepath.Join("users", userID.String(), "items", itemID.String(), photoID.String(), "thumbnail.jpg")
}
