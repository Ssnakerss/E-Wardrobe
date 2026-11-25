package items

import (
	"context"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// ApplyCreate applies a create mutation from the sync protocol.
func ApplyCreate(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change, svc *Service) error {
	payload := mutation.Payload
	if payload == nil {
		return fmt.Errorf("empty payload")
	}

	// Extract fields from payload
	name, _ := payload["name"].(string)
	if name == "" {
		return fmt.Errorf("missing required field: name")
	}

	categoryID, _ := payload["category_id"].(uuid.UUID)

	brand, _ := payload["brand"].(string)
	var brandPtr *string
	if brand != "" {
		brandPtr = &brand
	}

	description, _ := payload["description"].(string)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	purchaseYear, ok := payload["purchase_year"].(float64)
	var purchaseYearPtr *int16
	if ok {
		v := int16(purchaseYear)
		purchaseYearPtr = &v
	}

	purchasePrice, ok := payload["purchase_price"].(float64)
	var purchasePricePtr *int64
	if ok {
		v := int64(purchasePrice)
		purchasePricePtr = &v
	}

	priceCurrency, _ := payload["price_currency"].(string)
	var priceCurrencyPtr *string
	if priceCurrency != "" {
		priceCurrencyPtr = &priceCurrency
	}

	purchasePlace, _ := payload["purchase_place"].(string)
	var purchasePlacePtr *string
	if purchasePlace != "" {
		purchasePlacePtr = &purchasePlace
	}

	conditionRating, ok := payload["condition_rating"].(float64)
	var conditionRatingPtr *int16
	if ok {
		v := int16(conditionRating)
		conditionRatingPtr = &v
	}

	preferenceRating, ok := payload["preference_rating"].(float64)
	var preferenceRatingPtr *int16
	if ok {
		v := int16(preferenceRating)
		preferenceRatingPtr = &v
	}

	// Parse last_worn_on
	var lastWornOnPtr *time.Time
	if lwo, ok := payload["last_worn_on"].(string); ok && lwo != "" {
		t, err := time.Parse("2006-01-02", lwo)
		if err == nil {
			lastWornOnPtr = &t
		}
	}

	storageLocation, _ := payload["storage_location"].(string)
	var storageLocationPtr *string
	if storageLocation != "" {
		storageLocationPtr = &storageLocation
	}

	season, _ := payload["season"].(string)
	var seasonPtr *string
	if season != "" {
		seasonPtr = &season
	}

	color, _ := payload["color"].(string)
	var colorPtr *string
	if color != "" {
		colorPtr = &color
	}

	size, _ := payload["size"].(string)
	var sizePtr *string
	if size != "" {
		sizePtr = &size
	}

	status, _ := payload["status"].(string)
	if status == "" {
		status = "in_use"
	}

	input := ItemInput{
		CategoryID:       categoryID,
		Name:             name,
		Brand:            brandPtr,
		Description:      descriptionPtr,
		PurchaseYear:     purchaseYearPtr,
		PurchasePrice:    purchasePricePtr,
		PriceCurrency:    priceCurrencyPtr,
		PurchasePlace:    purchasePlacePtr,
		ConditionRating:  conditionRatingPtr,
		PreferenceRating: preferenceRatingPtr,
		LastWornOn:       lastWornOnPtr,
		StorageLocation:  storageLocationPtr,
		Season:           seasonPtr,
		Color:            colorPtr,
		Size:             sizePtr,
		Status:           status,
	}

	_, err := svc.Create(ctx, tx, userID, input)
	if err != nil {
		return fmt.Errorf("apply create: %w", err)
	}

	return nil
}

// ApplyUpdate applies an update mutation from the sync protocol.
func ApplyUpdate(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change, svc *Service) error {
	payload := mutation.Payload
	if payload == nil {
		return fmt.Errorf("empty payload")
	}

	entityID, _ := payload["entity_id"].(uuid.UUID)
	if entityID == uuid.Nil {
		entityID = mutation.EntityID
	}

	// Extract base_version from payload or use EntityVersion
	baseVersion := int64(0)
	if bv, ok := payload["base_version"].(float64); ok {
		baseVersion = int64(bv)
	} else if mutation.EntityVersion > 0 {
		baseVersion = mutation.EntityVersion - 1
	}

	// Build patch from payload
	var patch ItemPatch
	patch.ID = entityID
	patch.BaseVersion = baseVersion

	if name, ok := payload["name"].(string); ok {
		patch.Name = &name
	}
	if categoryID, ok := payload["category_id"].(uuid.UUID); ok {
		patch.CategoryID = &categoryID
	}
	if brand, ok := payload["brand"].(string); ok {
		patch.Brand = &brand
	}
	if description, ok := payload["description"].(string); ok {
		patch.Description = &description
	}
	if purchaseYear, ok := payload["purchase_year"].(float64); ok {
		v := int16(purchaseYear)
		patch.PurchaseYear = &v
	}
	if purchasePrice, ok := payload["purchase_price"].(float64); ok {
		v := int64(purchasePrice)
		patch.PurchasePrice = &v
	}
	if priceCurrency, ok := payload["price_currency"].(string); ok {
		patch.PriceCurrency = &priceCurrency
	}
	if purchasePlace, ok := payload["purchase_place"].(string); ok {
		patch.PurchasePlace = &purchasePlace
	}
	if conditionRating, ok := payload["condition_rating"].(float64); ok {
		v := int16(conditionRating)
		patch.ConditionRating = &v
	}
	if preferenceRating, ok := payload["preference_rating"].(float64); ok {
		v := int16(preferenceRating)
		patch.PreferenceRating = &v
	}
	if season, ok := payload["season"].(string); ok {
		patch.Season = &season
	}
	if color, ok := payload["color"].(string); ok {
		patch.Color = &color
	}
	if size, ok := payload["size"].(string); ok {
		patch.Size = &size
	}
	if status, ok := payload["status"].(string); ok {
		patch.Status = &status
	}
	if storageLocation, ok := payload["storage_location"].(string); ok {
		patch.StorageLocation = &storageLocation
	}

	_, err := svc.Update(ctx, tx, userID, patch)
	if err != nil {
		return fmt.Errorf("apply update: %w", err)
	}

	return nil
}

// ApplyDelete applies a delete mutation from the sync protocol.
func ApplyDelete(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change, svc *Service) error {
	// Verify ownership
	_, err := svc.GetByID(ctx, tx, mutation.EntityID, userID)
	if err != nil {
		return fmt.Errorf("item not found or not owned: %w", err)
	}

	// Get current version from payload or use EntityVersion
	version := int64(1)
	if v, ok := mutation.Payload["version"].(float64); ok {
		version = int64(v)
	} else if mutation.EntityVersion > 0 {
		version = mutation.EntityVersion
	}

	err = svc.SoftDelete(ctx, tx, mutation.EntityID, userID, version)
	if err != nil {
		return fmt.Errorf("apply delete: %w", err)
	}

	return nil
}

// ApplyArchive applies an archive mutation from the sync protocol.
func ApplyArchive(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change, svc *Service) error {
	version := int64(1)
	if v, ok := mutation.Payload["version"].(float64); ok {
		version = int64(v)
	} else if mutation.EntityVersion > 0 {
		version = mutation.EntityVersion
	}

	_, err := svc.Archive(ctx, tx, mutation.EntityID, userID, version)
	if err != nil {
		return fmt.Errorf("apply archive: %w", err)
	}

	return nil
}

// ApplyUnarchive applies an unarchive mutation from the sync protocol.
func ApplyUnarchive(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change, svc *Service) error {
	version := int64(1)
	if v, ok := mutation.Payload["version"].(float64); ok {
		version = int64(v)
	} else if mutation.EntityVersion > 0 {
		version = mutation.EntityVersion
	}

	_, err := svc.Unarchive(ctx, tx, mutation.EntityID, userID, version)
	if err != nil {
		return fmt.Errorf("apply unarchive: %w", err)
	}

	return nil
}

// SnapshotItem creates a full snapshot of an item for sync.
func SnapshotItem(item Item) map[string]any {
	payload := map[string]any{
		"entity_id":      item.ID,
		"category_id":    item.CategoryID,
		"name":           item.Name,
		"brand":          item.Brand,
		"description":    item.Description,
		"purchase_year":  item.PurchaseYear,
		"purchase_price": item.PurchasePrice,
		"price_currency": item.PriceCurrency,
		"purchase_place": item.PurchasePlace,
		"condition_rating": item.ConditionRating,
		"preference_rating": item.PreferenceRating,
		"storage_location": item.StorageLocation,
		"season":         item.Season,
		"color":          item.Color,
		"size":           item.Size,
		"status":         item.Status,
		"version":        item.Version,
	}

	if item.LastWornOn != nil {
		payload["last_worn_on"] = item.LastWornOn.Format("2006-01-02")
	}

	return payload
}

// GetItemForSync retrieves an item with its version for sync operations.
func GetItemForSync(ctx context.Context, tx database.Tx, repo *Repository, id, userID uuid.UUID) (Item, error) {
	return repo.GetByID(ctx, tx, id, userID)
}

// CheckProcessedMutation checks if a mutation has already been processed.
func CheckProcessedMutation(ctx context.Context, tx database.Tx, mutationID uuid.UUID, userID uuid.UUID) (bool, error) {
	q := sqlc.NewFromTx(tx)

	_, err := q.GetProcessedMutation(ctx, sqlc.GetProcessedMutationParams{
		UserID:     userID,
		MutationID: mutationID,
	})
	if err != nil {
		if err == sqlc.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("check processed mutation: %w", err)
	}

	return true, nil
}

// RecordProcessedMutation marks a mutation as processed.
func RecordProcessedMutation(ctx context.Context, tx database.Tx, mutationID uuid.UUID, userID uuid.UUID, entity string, entityID uuid.UUID, result map[string]any) error {
	q := sqlc.NewFromTx(tx)

	err := q.CreateProcessedMutation(ctx, sqlc.CreateProcessedMutationParams{
		UserID:      userID,
		MutationID:  mutationID,
		Entity:      entity,
		EntityID:    entityID,
		Result:      result,
	})
	if err != nil {
		return fmt.Errorf("record processed mutation: %w", err)
	}

	return nil
}
