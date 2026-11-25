package photos

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/storage"
	"github.com/google/uuid"
)

// UploadPipeline handles the full photo upload lifecycle.
type UploadPipeline struct {
	repo    *Repository
	store   storage.PhotoStorage
	journal *changejournal.Writer
	pool    *database.Pool
}

// NewUploadPipeline creates a new upload pipeline.
func NewUploadPipeline(repo *Repository, store storage.PhotoStorage, journal *changejournal.Writer, pool *database.Pool) *UploadPipeline {
	return &UploadPipeline{repo: repo, store: store, journal: journal, pool: pool}
}

// UploadPhoto performs a full upload pipeline for web clients.
// It validates, normalizes, creates thumbnails, and saves to storage.
func (p *UploadPipeline) UploadPhoto(ctx context.Context, userID, itemID uuid.UUID, file io.Reader, sortOrder int32) (Photo, error) {
	// Read and validate file size
	data, err := ReadLimited(file, MaxUploadSize)
	if err != nil {
		return Photo{}, fmt.Errorf("read file: %w", err)
	}

	// Validate image content
	if err := ValidateImage(data); err != nil {
		return Photo{}, fmt.Errorf("validate image: %w", err)
	}

	// Decode to get dimensions
	width, height, err := DecodeImage(data)
	if err != nil {
		return Photo{}, fmt.Errorf("decode image: %w", err)
	}

	// Normalize to JPEG
	normalized, err := NormalizeJPEG(data)
	if err != nil {
		return Photo{}, fmt.Errorf("normalize JPEG: %w", err)
	}

	// Create thumbnail
	thumbnail, err := CreateThumbnail(normalized, MaxThumbnailSide)
	if err != nil {
		return Photo{}, fmt.Errorf("create thumbnail: %w", err)
	}

	// Compute content hash
	hash := sha256.Sum256(normalized)
	contentHash := fmt.Sprintf("%x", hash)

	// Generate photo ID
	photoID := uuid.New()

	// Build storage keys
	keyBuilder := &StorageKeyBuilder{
		UserID:  userID,
		ItemID:  itemID,
		PhotoID: photoID,
	}
	objectKey := keyBuilder.OriginalKey()
	thumbnailKey := keyBuilder.ThumbnailKey()

	// Save to storage (outside transaction)
	if err := p.store.Save(ctx, objectKey, bytes.NewReader(normalized)); err != nil {
		return Photo{}, fmt.Errorf("save original: %w", err)
	}

	if err := p.store.Save(ctx, thumbnailKey, bytes.NewReader(thumbnail)); err != nil {
		// Compensating: delete original
		_ = p.store.Delete(ctx, objectKey)
		return Photo{}, fmt.Errorf("save thumbnail: %w", err)
	}

	// Create photo metadata in transaction
	photo := Photo{
		ID:            photoID,
		ItemID:        itemID,
		UserID:        userID,
		SortOrder:     sortOrder,
		Status:        PhotoStatusReady,
		ObjectKey:     &objectKey,
		ThumbnailKey:  &thumbnailKey,
		MimeType:      strPtr("image/jpeg"),
		SizeBytes:     int64Ptr(int64(len(normalized))),
		Width:         int32Ptr(int32(width)),
		Height:        int32Ptr(int32(height)),
		ContentHash:   &contentHash,
		Version:       1,
		CreatedAt:     txCreatedAt(),
		UpdatedAt:     txCreatedAt(),
	}

	photo, err = p.repo.Create(ctx, p.pool, photo)
	if err != nil {
		// Compensating: delete stored files
		_ = p.store.Delete(ctx, objectKey)
		_ = p.store.Delete(ctx, thumbnailKey)
		return Photo{}, fmt.Errorf("create photo: %w", err)
	}

	// Write to change journal
	err = p.journal.WriteRecord(ctx, p.pool, changejournal.Change{
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

// UploadContent performs an idempotent content upload for mobile clients.
// The photo metadata must already exist with status awaiting_upload.
func (p *UploadPipeline) UploadContent(ctx context.Context, userID uuid.UUID, photoID uuid.UUID, file io.Reader) (Photo, error) {
	// Read and validate file size
	data, err := ReadLimited(file, MaxUploadSize)
	if err != nil {
		return Photo{}, fmt.Errorf("read file: %w", err)
	}

	// Validate image content
	if err := ValidateImage(data); err != nil {
		return Photo{}, fmt.Errorf("validate image: %w", err)
	}

	// Get existing photo
	photo, err := p.repo.GetByID(ctx, p.pool, photoID, userID)
	if err != nil {
		return Photo{}, fmt.Errorf("photo not found or not owned: %w", err)
	}

	// Check if already ready (idempotent)
	if photo.Status == PhotoStatusReady {
		return photo, nil
	}

	// Decode to get dimensions
	width, height, err := DecodeImage(data)
	if err != nil {
		return Photo{}, fmt.Errorf("decode image: %w", err)
	}

	// Normalize to JPEG
	normalized, err := NormalizeJPEG(data)
	if err != nil {
		return Photo{}, fmt.Errorf("normalize JPEG: %w", err)
	}

	// Create thumbnail
	thumbnail, err := CreateThumbnail(normalized, MaxThumbnailSide)
	if err != nil {
		return Photo{}, fmt.Errorf("create thumbnail: %w", err)
	}

	// Compute content hash
	hash := sha256.Sum256(normalized)
	contentHash := fmt.Sprintf("%x", hash)

	// Use existing keys or generate new ones
	objectKey := photo.ObjectKey
	thumbnailKey := photo.ThumbnailKey
	if objectKey == nil {
		keyBuilder := &StorageKeyBuilder{
			UserID:  userID,
			ItemID:  photo.ItemID,
			PhotoID: photoID,
		}
		objectKey = strPtr(keyBuilder.OriginalKey())
		thumbnailKey = strPtr(keyBuilder.ThumbnailKey())
	}

	// Save to storage (outside transaction)
	if err := p.store.Save(ctx, *objectKey, bytes.NewReader(normalized)); err != nil {
		return Photo{}, fmt.Errorf("save original: %w", err)
	}

	if err := p.store.Save(ctx, *thumbnailKey, bytes.NewReader(thumbnail)); err != nil {
		_ = p.store.Delete(ctx, *objectKey)
		return Photo{}, fmt.Errorf("save thumbnail: %w", err)
	}

	// Update photo metadata in transaction
	patch := PhotoPatch{
		ID:            photoID,
		BaseVersion:   photo.Version,
		Status:        photoStatusPtr(PhotoStatusReady),
		ObjectKey:     objectKey,
		ThumbnailKey:  thumbnailKey,
		MimeType:      strPtr("image/jpeg"),
		SizeBytes:     int64Ptr(int64(len(normalized))),
		Width:         int32Ptr(int32(width)),
		Height:        int32Ptr(int32(height)),
		ContentHash:   &contentHash,
	}

	updated, err := p.repo.Update(ctx, p.pool, patch)
	if err != nil {
		_ = p.store.Delete(ctx, *objectKey)
		_ = p.store.Delete(ctx, *thumbnailKey)
		return Photo{}, fmt.Errorf("update photo: %w", err)
	}

	// Write to change journal
	err = p.journal.WriteRecord(ctx, p.pool, changejournal.Change{
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

// Helper functions
func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func photoStatusPtr(s PhotoStatus) *PhotoStatus {
	return &s
}

func txCreatedAt() time.Time {
	return time.Now().UTC()
}
