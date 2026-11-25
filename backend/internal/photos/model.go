package photos

import (
	"time"

	"github.com/google/uuid"
)

// PhotoStatus represents the status of a photo.
type PhotoStatus string

const (
	PhotoStatusAwaitingUpload PhotoStatus = "awaiting_upload"
	PhotoStatusProcessing     PhotoStatus = "processing"
	PhotoStatusReady          PhotoStatus = "ready"
	PhotoStatusFailed         PhotoStatus = "failed"
)

// Photo represents a photo attached to a wardrobe item.
type Photo struct {
	ID             uuid.UUID
	ItemID         uuid.UUID
	UserID         uuid.UUID
	SortOrder      int32
	Status         PhotoStatus
	ObjectKey      *string
	ThumbnailKey   *string
	MimeType       *string
	SizeBytes      *int64
	Width          *int32
	Height         *int32
	ContentHash    *string
	ProcessingError *string
	Version        int64
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PhotoInput represents input for creating a photo metadata record.
type PhotoInput struct {
	ItemID   uuid.UUID
	SortOrder int32
}

// PhotoPatch represents input for updating a photo.
type PhotoPatch struct {
	ID             uuid.UUID
	BaseVersion    int64
	Status         *PhotoStatus
	ObjectKey      *string
	ThumbnailKey   *string
	MimeType       *string
	SizeBytes      *int64
	Width          *int32
	Height         *int32
	ContentHash    *string
	ProcessingError *string
}

// PhotoOrder represents a requested sort order change.
type PhotoOrder struct {
	ID        uuid.UUID
	SortOrder int32
}

// PhotoContent represents decoded image data with metadata.
type PhotoContent struct {
	Data      []byte
	MimeType  string
	Width     int
	Height    int
	ContentHash string
}

// StorageKeyBuilder builds storage keys for photo content.
type StorageKeyBuilder struct {
	UserID   uuid.UUID
	ItemID   uuid.UUID
	PhotoID  uuid.UUID
}

// OriginalKey returns the storage key for the original image.
func (k *StorageKeyBuilder) OriginalKey() string {
	return "users/" + k.UserID.String() + "/items/" + k.ItemID.String() + "/" + k.PhotoID.String() + "/original.jpg"
}

// ThumbnailKey returns the storage key for the thumbnail.
func (k *StorageKeyBuilder) ThumbnailKey() string {
	return "users/" + k.UserID.String() + "/items/" + k.ItemID.String() + "/" + k.PhotoID.String() + "/thumbnail.jpg"
}
