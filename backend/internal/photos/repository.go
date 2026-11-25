package photos

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// Repository handles database operations for photos.
type Repository struct {
	pool *database.Pool
}

// NewRepository creates a new photos repository.
func NewRepository(pool *database.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new photo metadata record.
func (r *Repository) Create(ctx context.Context, exec database.Executor, photo Photo) (Photo, error) {
	q := sqlc.NewFromExecutor(exec)

	var objectKeySQL sql.NullString
	if photo.ObjectKey != nil {
		objectKeySQL = sql.NullString{String: *photo.ObjectKey, Valid: true}
	}

	var thumbnailKeySQL sql.NullString
	if photo.ThumbnailKey != nil {
		thumbnailKeySQL = sql.NullString{String: *photo.ThumbnailKey, Valid: true}
	}

	var mimeTypeSQL sql.NullString
	if photo.MimeType != nil {
		mimeTypeSQL = sql.NullString{String: *photo.MimeType, Valid: true}
	}

	var sizeBytesSQL sql.NullInt64
	if photo.SizeBytes != nil {
		sizeBytesSQL = sql.NullInt64{Int64: *photo.SizeBytes, Valid: true}
	}

	var widthSQL sql.NullInt32
	if photo.Width != nil {
		widthSQL = sql.NullInt32{Int32: *photo.Width, Valid: true}
	}

	var heightSQL sql.NullInt32
	if photo.Height != nil {
		heightSQL = sql.NullInt32{Int32: *photo.Height, Valid: true}
	}

	p, err := q.CreatePhoto(ctx, sqlc.CreatePhotoParams{
		ID:           photo.ID,
		ItemID:       photo.ItemID,
		UserID:       photo.UserID,
		SortOrder:    photo.SortOrder,
		Status:       string(photo.Status),
		ObjectKey:    objectKeySQL,
		ThumbnailKey: thumbnailKeySQL,
		MimeType:     mimeTypeSQL,
		SizeBytes:    sizeBytesSQL,
		Width:        widthSQL,
		Height:       heightSQL,
		Version:      photo.Version,
	})
	if err != nil {
		return photo, fmt.Errorf("create photo: %w", err)
	}

	return toPhoto(p), nil
}

// GetByID returns a photo by ID with ownership check.
func (r *Repository) GetByID(ctx context.Context, exec database.Executor, id, userID uuid.UUID) (Photo, error) {
	q := sqlc.NewFromExecutor(exec)

	p, err := q.GetPhoto(ctx, sqlc.GetPhotoParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return Photo{}, fmt.Errorf("get photo: %w", err)
	}

	return toPhoto(p), nil
}

// GetByItem returns all photos for an item.
func (r *Repository) GetByItem(ctx context.Context, exec database.Executor, itemID, userID uuid.UUID) ([]Photo, error) {
	q := sqlc.NewFromExecutor(exec)

	photos, err := q.GetPhotoByItem(ctx, sqlc.GetPhotoByItemParams{
		ItemID: itemID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get photos by item: %w", err)
	}

	result := make([]Photo, len(photos))
	for i, p := range photos {
		result[i] = toPhoto(p)
	}

	return result, nil
}

// Update updates a photo metadata record.
func (r *Repository) Update(ctx context.Context, exec database.Executor, patch PhotoPatch) (Photo, error) {
	q := sqlc.NewFromExecutor(exec)

	var statusSQL sql.NullString
	if patch.Status != nil {
		statusSQL = sql.NullString{String: string(*patch.Status), Valid: true}
	}

	var objectKeySQL sql.NullString
	if patch.ObjectKey != nil {
		objectKeySQL = sql.NullString{String: *patch.ObjectKey, Valid: true}
	}

	var thumbnailKeySQL sql.NullString
	if patch.ThumbnailKey != nil {
		thumbnailKeySQL = sql.NullString{String: *patch.ThumbnailKey, Valid: true}
	}

	var mimeTypeSQL sql.NullString
	if patch.MimeType != nil {
		mimeTypeSQL = sql.NullString{String: *patch.MimeType, Valid: true}
	}

	var sizeBytesSQL sql.NullInt64
	if patch.SizeBytes != nil {
		sizeBytesSQL = sql.NullInt64{Int64: *patch.SizeBytes, Valid: true}
	}

	var widthSQL sql.NullInt32
	if patch.Width != nil {
		widthSQL = sql.NullInt32{Int32: *patch.Width, Valid: true}
	}

	var heightSQL sql.NullInt32
	if patch.Height != nil {
		heightSQL = sql.NullInt32{Int32: *patch.Height, Valid: true}
	}

	p, err := q.UpdatePhoto(ctx, sqlc.UpdatePhotoParams{
		ID:             patch.ID,
		UserID:         patch.ID,
		Status:         statusSQL,
		ObjectKey:      objectKeySQL,
		ThumbnailKey:   thumbnailKeySQL,
		MimeType:       mimeTypeSQL,
		SizeBytes:      sizeBytesSQL,
		Width:          widthSQL,
		Height:         heightSQL,
		Version:        patch.BaseVersion,
	})
	if err != nil {
		return Photo{}, fmt.Errorf("update photo: %w", err)
	}

	return toPhoto(p), nil
}

// SoftDelete marks a photo as deleted.
func (r *Repository) SoftDelete(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) error {
	q := sqlc.NewFromExecutor(exec)

	_, err := q.SoftDeletePhoto(ctx, sqlc.SoftDeletePhotoParams{
		ID:      id,
		UserID:  userID,
		Version: version,
	})
	if err != nil {
		return fmt.Errorf("soft delete photo: %w", err)
	}

	return nil
}

// UpdateSortOrder updates the sort order of a photo.
func (r *Repository) UpdateSortOrder(ctx context.Context, exec database.Executor, id, userID uuid.UUID, sortOrder int32, version int64) (Photo, error) {
	q := sqlc.NewFromExecutor(exec)

	p, err := q.UpdatePhotoSortOrder(ctx, sqlc.UpdatePhotoSortOrderParams{
		ID:        id,
		UserID:    userID,
		SortOrder: sortOrder,
		Version:   version,
	})
	if err != nil {
		return Photo{}, fmt.Errorf("update photo sort order: %w", err)
	}

	return toPhoto(p), nil
}

// toPhoto converts sqlc.Photo to domain Photo.
func toPhoto(p sqlc.Photo) Photo {
	return Photo{
		ID:             p.ID,
		ItemID:         p.ItemID,
		UserID:         p.UserID,
		SortOrder:      p.SortOrder,
		Status:         PhotoStatus(p.Status),
		ObjectKey:      nullableStringFromSQL(p.ObjectKey),
		ThumbnailKey:   nullableStringFromSQL(p.ThumbnailKey),
		MimeType:       nullableStringFromSQL(p.MimeType),
		SizeBytes:      nullableInt64FromSQL(p.SizeBytes),
		Width:          nullableInt32FromSQL(p.Width),
		Height:         nullableInt32FromSQL(p.Height),
		ContentHash:    nullableStringFromSQL(p.ContentHash),
		ProcessingError: nullableStringFromSQL(p.ProcessingError),
		Version:        p.Version,
		DeletedAt:      nullableTimeFromSQL(p.DeletedAt),
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func nullableStringFromSQL(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func nullableInt64FromSQL(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func nullableInt32FromSQL(n sql.NullInt32) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}

func nullableTimeFromSQL(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
