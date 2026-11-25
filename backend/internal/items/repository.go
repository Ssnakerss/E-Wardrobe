package items

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// Repository handles database operations for items.
type Repository struct {
	pool *database.Pool
}

// NewRepository creates a new items repository.
func NewRepository(pool *database.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new item.
func (r *Repository) Create(ctx context.Context, exec database.Executor, item Item) (Item, error) {
	q := sqlc.NewFromExecutor(exec)

	var brandSQL sql.NullString
	if item.Brand != nil {
		brandSQL = sql.NullString{String: *item.Brand, Valid: true}
	}

	var descriptionSQL sql.NullString
	if item.Description != nil {
		descriptionSQL = sql.NullString{String: *item.Description, Valid: true}
	}

	var purchaseYearSQL sql.NullInt16
	if item.PurchaseYear != nil {
		purchaseYearSQL = sql.NullInt16{Int16: *item.PurchaseYear, Valid: true}
	}

	var purchasePriceSQL sql.NullInt64
	if item.PurchasePrice != nil {
		purchasePriceSQL = sql.NullInt64{Int64: *item.PurchasePrice, Valid: true}
	}

	var priceCurrencySQL sql.NullString
	if item.PriceCurrency != nil {
		priceCurrencySQL = sql.NullString{String: *item.PriceCurrency, Valid: true}
	}

	var purchasePlaceSQL sql.NullString
	if item.PurchasePlace != nil {
		purchasePlaceSQL = sql.NullString{String: *item.PurchasePlace, Valid: true}
	}

	var conditionRatingSQL sql.NullInt16
	if item.ConditionRating != nil {
		conditionRatingSQL = sql.NullInt16{Int16: *item.ConditionRating, Valid: true}
	}

	var preferenceRatingSQL sql.NullInt16
	if item.PreferenceRating != nil {
		preferenceRatingSQL = sql.NullInt16{Int16: *item.PreferenceRating, Valid: true}
	}

	var lastWornOnSQL sql.NullTime
	if item.LastWornOn != nil {
		lastWornOnSQL = sql.NullTime{Time: *item.LastWornOn, Valid: true}
	}

	var storageLocationSQL sql.NullString
	if item.StorageLocation != nil {
		storageLocationSQL = sql.NullString{String: *item.StorageLocation, Valid: true}
	}

	var seasonSQL sql.NullString
	if item.Season != nil {
		seasonSQL = sql.NullString{String: *item.Season, Valid: true}
	}

	var colorSQL sql.NullString
	if item.Color != nil {
		colorSQL = sql.NullString{String: *item.Color, Valid: true}
	}

	var sizeSQL sql.NullString
	if item.Size != nil {
		sizeSQL = sql.NullString{String: *item.Size, Valid: true}
	}

	i, err := q.CreateItem(ctx, sqlc.CreateItemParams{
		ID:               item.ID,
		UserID:           item.UserID,
		CategoryID:       item.CategoryID,
		Name:             item.Name,
		Brand:            brandSQL,
		Description:      descriptionSQL,
		PurchaseYear:     purchaseYearSQL,
		PurchasePrice:    purchasePriceSQL,
		PriceCurrency:    priceCurrencySQL,
		PurchasePlace:    purchasePlaceSQL,
		ConditionRating:  conditionRatingSQL,
		PreferenceRating: preferenceRatingSQL,
		LastWornOn:       lastWornOnSQL,
		StorageLocation:  storageLocationSQL,
		Season:           seasonSQL,
		Color:            colorSQL,
		Size:             sizeSQL,
		Status:           item.Status,
		Version:          item.Version,
	})
	if err != nil {
		return item, fmt.Errorf("create item: %w", err)
	}

	return toItem(i), nil
}

// GetByID returns an item by ID with ownership check.
func (r *Repository) GetByID(ctx context.Context, exec database.Executor, id, userID uuid.UUID) (Item, error) {
	q := sqlc.NewFromExecutor(exec)

	i, err := q.GetItem(ctx, sqlc.GetItemParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return Item{}, fmt.Errorf("get item: %w", err)
	}

	return toItem(i), nil
}

// Update updates an item with version check.
func (r *Repository) Update(ctx context.Context, exec database.Executor, patch ItemPatch) (Item, error) {
	q := sqlc.NewFromExecutor(exec)

	var categoryIDSQL *uuid.UUID
	if patch.CategoryID != nil {
		categoryIDSQL = patch.CategoryID
	}

	var nameSQL sql.NullString
	if patch.Name != nil {
		nameSQL = sql.NullString{String: *patch.Name, Valid: true}
	}

	var brandSQL sql.NullString
	if patch.Brand != nil {
		brandSQL = sql.NullString{String: *patch.Brand, Valid: true}
	}

	var descriptionSQL sql.NullString
	if patch.Description != nil {
		descriptionSQL = sql.NullString{String: *patch.Description, Valid: true}
	}

	var purchaseYearSQL sql.NullInt16
	if patch.PurchaseYear != nil {
		purchaseYearSQL = sql.NullInt16{Int16: *patch.PurchaseYear, Valid: true}
	}

	var purchasePriceSQL sql.NullInt64
	if patch.PurchasePrice != nil {
		purchasePriceSQL = sql.NullInt64{Int64: *patch.PurchasePrice, Valid: true}
	}

	var priceCurrencySQL sql.NullString
	if patch.PriceCurrency != nil {
		priceCurrencySQL = sql.NullString{String: *patch.PriceCurrency, Valid: true}
	}

	var purchasePlaceSQL sql.NullString
	if patch.PurchasePlace != nil {
		purchasePlaceSQL = sql.NullString{String: *patch.PurchasePlace, Valid: true}
	}

	var conditionRatingSQL sql.NullInt16
	if patch.ConditionRating != nil {
		conditionRatingSQL = sql.NullInt16{Int16: *patch.ConditionRating, Valid: true}
	}

	var preferenceRatingSQL sql.NullInt16
	if patch.PreferenceRating != nil {
		preferenceRatingSQL = sql.NullInt16{Int16: *patch.PreferenceRating, Valid: true}
	}

	var lastWornOnSQL sql.NullTime
	if patch.LastWornOn != nil {
		lastWornOnSQL = sql.NullTime{Time: *patch.LastWornOn, Valid: true}
	}

	var storageLocationSQL sql.NullString
	if patch.StorageLocation != nil {
		storageLocationSQL = sql.NullString{String: *patch.StorageLocation, Valid: true}
	}

	var seasonSQL sql.NullString
	if patch.Season != nil {
		seasonSQL = sql.NullString{String: *patch.Season, Valid: true}
	}

	var colorSQL sql.NullString
	if patch.Color != nil {
		colorSQL = sql.NullString{String: *patch.Color, Valid: true}
	}

	var sizeSQL sql.NullString
	if patch.Size != nil {
		sizeSQL = sql.NullString{String: *patch.Size, Valid: true}
	}

	var statusSQL sql.NullString
	if patch.Status != nil {
		statusSQL = sql.NullString{String: *patch.Status, Valid: true}
	}

	i, err := q.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:               patch.ID,
		UserID:           patch.ID,
		CategoryID:       categoryIDSQL,
		Name:             nameSQL,
		Brand:            brandSQL,
		Description:      descriptionSQL,
		PurchaseYear:     purchaseYearSQL,
		PurchasePrice:    purchasePriceSQL,
		PriceCurrency:    priceCurrencySQL,
		PurchasePlace:    purchasePlaceSQL,
		ConditionRating:  conditionRatingSQL,
		PreferenceRating: preferenceRatingSQL,
		LastWornOn:       lastWornOnSQL,
		StorageLocation:  storageLocationSQL,
		Season:           seasonSQL,
		Color:            colorSQL,
		Size:             sizeSQL,
		Status:           statusSQL,
		Version:          patch.BaseVersion,
	})
	if err != nil {
		return Item{}, fmt.Errorf("update item: %w", err)
	}

	return toItem(i), nil
}

// SoftDelete marks an item as deleted.
func (r *Repository) SoftDelete(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) error {
	q := sqlc.NewFromExecutor(exec)

	_, err := q.SoftDeleteItem(ctx, sqlc.SoftDeleteItemParams{
		ID:      id,
		UserID:  userID,
		Version: version,
	})
	if err != nil {
		return fmt.Errorf("soft delete item: %w", err)
	}

	return nil
}

// Archive marks an item as archived.
func (r *Repository) Archive(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) (Item, error) {
	q := sqlc.NewFromExecutor(exec)

	i, err := q.ArchiveItem(ctx, sqlc.ArchiveItemParams{
		ID:      id,
		UserID:  userID,
		Version: version,
	})
	if err != nil {
		return Item{}, fmt.Errorf("archive item: %w", err)
	}

	return toItem(i), nil
}

// Unarchive removes the archive flag from an item.
func (r *Repository) Unarchive(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) (Item, error) {
	q := sqlc.NewFromExecutor(exec)

	i, err := q.UnarchiveItem(ctx, sqlc.UnarchiveItemParams{
		ID:      id,
		UserID:  userID,
		Version: version,
	})
	if err != nil {
		return Item{}, fmt.Errorf("unarchive item: %w", err)
	}

	return toItem(i), nil
}

// List returns items matching the options with filtering, sorting and pagination.
func (r *Repository) List(ctx context.Context, exec database.Executor, opts ListOptions) ([]Item, error) {
	q := sqlc.NewFromExecutor(exec)

	var querySQL sql.NullString
	if opts.Query != nil {
		querySQL = sql.NullString{String: "%" + *opts.Query + "%", Valid: true}
	}

	var statusSQL sql.NullString
	if opts.Status != nil {
		statusSQL = sql.NullString{String: *opts.Status, Valid: true}
	}

	var seasonSQL sql.NullString
	if opts.Season != nil {
		seasonSQL = sql.NullString{String: *opts.Season, Valid: true}
	}

	var conditionSQL sql.NullInt16
	if opts.Condition != nil {
		conditionSQL = sql.NullInt16{Int16: *opts.Condition, Valid: true}
	}

	items, err := q.ListItems(ctx, sqlc.ListItemsParams{
		UserID:     opts.UserID,
		CategoryID: opts.CategoryID,
		Query:      querySQL,
		Status:     statusSQL,
		Season:     seasonSQL,
		Condition:  conditionSQL,
		Archived:   opts.Archived,
		Limit:      opts.Limit,
		Offset:     opts.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	result := make([]Item, len(items))
	for i, item := range items {
		result[i] = toItem(item)
	}

	return result, nil
}

// CountItems returns the count of items matching the options.
func (r *Repository) CountItems(ctx context.Context, exec database.Executor, opts ListOptions) (int64, error) {
	q := sqlc.NewFromExecutor(exec)

	var querySQL sql.NullString
	if opts.Query != nil {
		querySQL = sql.NullString{String: "%" + *opts.Query + "%", Valid: true}
	}

	var statusSQL sql.NullString
	if opts.Status != nil {
		statusSQL = sql.NullString{String: *opts.Status, Valid: true}
	}

	count, err := q.CountItems(ctx, sqlc.CountItemsParams{
		UserID:     opts.UserID,
		CategoryID: opts.CategoryID,
		Query:      querySQL,
		Status:     statusSQL,
		Archived:   opts.Archived,
	})
	if err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}

	return count, nil
}

// ListStorageLocationSuggestions returns distinct storage locations matching the query.
func (r *Repository) ListStorageLocationSuggestions(ctx context.Context, exec database.Executor, userID uuid.UUID, query string, limit int32) ([]string, error) {
	q := sqlc.NewFromExecutor(exec)

	suggestions, err := q.ListStorageLocationSuggestions(ctx, sqlc.ListStorageLocationSuggestionsParams{
		UserID: userID,
		Query:  query,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list storage location suggestions: %w", err)
	}

	return suggestions, nil
}

// toItem converts sqlc.Item to domain Item.
func toItem(i sqlc.Item) Item {
	return Item{
		ID:               i.ID,
		UserID:           i.UserID,
		CategoryID:       i.CategoryID,
		Name:             i.Name,
		Brand:            nullableStringFromSQL(i.Brand),
		Description:      nullableStringFromSQL(i.Description),
		PurchaseYear:     nullableInt16FromSQL(i.PurchaseYear),
		PurchasePrice:    nullableInt64FromSQL(i.PurchasePrice),
		PriceCurrency:    nullableStringFromSQL(i.PriceCurrency),
		PurchasePlace:    nullableStringFromSQL(i.PurchasePlace),
		ConditionRating:  nullableInt16FromSQL(i.ConditionRating),
		PreferenceRating: nullableInt16FromSQL(i.PreferenceRating),
		LastWornOn:       nullableTimeFromSQL(i.LastWornOn),
		StorageLocation:  nullableStringFromSQL(i.StorageLocation),
		Season:           nullableStringFromSQL(i.Season),
		Color:            nullableStringFromSQL(i.Color),
		Size:             nullableStringFromSQL(i.Size),
		Status:           i.Status,
		Version:          i.Version,
		ArchivedAt:       nullableTimeFromSQL(i.ArchivedAt),
		DeletedAt:        nullableTimeFromSQL(i.DeletedAt),
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
}

func nullableStringFromSQL(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func nullableInt16FromSQL(n sql.NullInt16) *int16 {
	if !n.Valid {
		return nil
	}
	return &n.Int16
}

func nullableInt64FromSQL(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func nullableTimeFromSQL(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
