package categories

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a category in the wardrobe.
type Category struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ParentID    *uuid.UUID
	TemplateKey *string
	Name        string
	SortOrder   int32
	Version     int64
	ArchivedAt  *time.Time
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CategoryInput represents input for creating a category.
type CategoryInput struct {
	ParentID    *uuid.UUID
	Name        string
	SortOrder   int32
}

// CategoryPatch represents input for updating a category.
type CategoryPatch struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	BaseVersion int64
	ParentID    *uuid.UUID
	Name        *string
	SortOrder   *int32
}

// ListOptions represents options for listing categories.
type ListOptions struct {
	UserID uuid.UUID
}
