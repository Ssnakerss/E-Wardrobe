package items

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a wardrobe item.
type Item struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	CategoryID      uuid.UUID
	Name            string
	Brand           *string
	Description     *string
	PurchaseYear    *int16
	PurchasePrice   *int64
	PriceCurrency   *string
	PurchasePlace   *string
	ConditionRating *int16
	PreferenceRating *int16
	LastWornOn      *time.Time
	StorageLocation *string
	Season          *string
	Color           *string
	Size            *string
	Status          string
	Version         int64
	ArchivedAt      *time.Time
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ItemInput represents input for creating an item.
type ItemInput struct {
	CategoryID       uuid.UUID
	Name             string
	Brand            *string
	Description      *string
	PurchaseYear     *int16
	PurchasePrice    *int64
	PriceCurrency    *string
	PurchasePlace    *string
	ConditionRating  *int16
	PreferenceRating *int16
	LastWornOn       *time.Time
	StorageLocation  *string
	Season           *string
	Color            *string
	Size             *string
	Status           string
}

// ItemPatch represents input for updating an item.
type ItemPatch struct {
	ID               uuid.UUID
	BaseVersion      int64
	CategoryID       *uuid.UUID
	Name             *string
	Brand            *string
	Description      *string
	PurchaseYear     *int16
	PurchasePrice    *int64
	PriceCurrency    *string
	PurchasePlace    *string
	ConditionRating  *int16
	PreferenceRating *int16
	LastWornOn       *time.Time
	StorageLocation  *string
	Season           *string
	Color            *string
	Size             *string
	Status           *string
}

// ListOptions represents options for filtering and listing items.
type ListOptions struct {
	UserID     uuid.UUID
	CategoryID *uuid.UUID
	Query      *string
	Status     *string
	Season     *string
	Condition  *int16
	Preference *int16
	Archived   bool
	Limit      int32
	Offset     int32
}

// StorageLocationSuggestionsOptions represents options for storage location suggestions.
type StorageLocationSuggestionsOptions struct {
	UserID uuid.UUID
	Query  string
	Limit  int32
}
