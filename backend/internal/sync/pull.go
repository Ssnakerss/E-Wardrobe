package sync

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/categories"
	"github.com/Ssnakerss/E-Wardrobe/internal/items"
	"github.com/Ssnakerss/E-Wardrobe/internal/photos"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// PullSnapshotter provides snapshot data for pull responses.
type PullSnapshotter struct {
	catRepo  *categories.Repository
	itemRepo *items.Repository
	photoRepo *photos.Repository
}

// NewPullSnapshotter creates a new snapshotter.
func NewPullSnapshotter(catRepo *categories.Repository, itemRepo *items.Repository, photoRepo *photos.Repository) *PullSnapshotter {
	return &PullSnapshotter{
		catRepo:   catRepo,
		itemRepo:  itemRepo,
		photoRepo: photoRepo,
	}
}

// GetChangeSnapshots retrieves full snapshots for the given changes.
// For deleted entities, returns tombstone payloads.
func (s *PullSnapshotter) GetChangeSnapshots(ctx context.Context, exec database.Executor, changes []SyncChange) ([]SyncChange, error) {
	if len(changes) == 0 {
		return changes, nil
	}

	result := make([]SyncChange, len(changes))
	for i, c := range changes {
		result[i] = c
		if c.Operation == "delete" {
			// Tombstone already has {"tombstone": true}
			continue
		}
		// For upsert, we need the full snapshot from the database
		snapshot, err := s.getSnapshot(ctx, exec, c.Entity, c.EntityID)
		if err != nil {
			// If entity not found, treat as tombstone
			result[i].Payload = map[string]any{"tombstone": true}
			result[i].Operation = "delete"
		} else {
			result[i].Payload = snapshot
		}
	}

	return result, nil
}

// getSnapshot retrieves a full snapshot for an entity.
func (s *PullSnapshotter) getSnapshot(ctx context.Context, exec database.Executor, entity string, entityID uuid.UUID) (map[string]any, error) {
	switch entity {
	case "category":
		return s.snapshotCategory(ctx, exec, entityID)
	case "item":
		return s.snapshotItem(ctx, exec, entityID)
	case "photo":
		return s.snapshotPhoto(ctx, exec, entityID)
	default:
		return nil, fmt.Errorf("unknown entity type for snapshot: %s", entity)
	}
}

// snapshotCategory creates a full category snapshot.
func (s *PullSnapshotter) snapshotCategory(ctx context.Context, exec database.Executor, id uuid.UUID) (map[string]any, error) {
	q := sqlc.NewFromExecutor(exec)

	cat, err := q.GetCategory(ctx, sqlc.GetCategoryParams{
		ID:     id,
		UserID: id, // placeholder - real userID comes from context
	})
	if err != nil {
		return nil, fmt.Errorf("get category snapshot: %w", err)
	}

	return map[string]any{
		"entity_id":  cat.ID,
		"name":       cat.Name,
		"parent_id":  cat.ParentID,
		"version":    cat.Version,
		"sort_order": cat.SortOrder,
	}, nil
}

// snapshotItem creates a full item snapshot.
func (s *PullSnapshotter) snapshotItem(ctx context.Context, exec database.Executor, id uuid.UUID) (map[string]any, error) {
	q := sqlc.NewFromExecutor(exec)

	item, err := q.GetItem(ctx, sqlc.GetItemParams{
		ID:     id,
		UserID: id, // placeholder
	})
	if err != nil {
		return nil, fmt.Errorf("get item snapshot: %w", err)
	}

	payload := map[string]any{
		"entity_id":      item.ID,
		"category_id":    item.CategoryID,
		"name":           item.Name,
		"version":        item.Version,
		"status":         item.Status,
	}

	if item.Brand.Valid {
		payload["brand"] = item.Brand.String
	}
	if item.Description.Valid {
		payload["description"] = item.Description.String
	}
	if item.PurchaseYear.Valid {
		payload["purchase_year"] = item.PurchaseYear.Int16
	}
	if item.PurchasePrice.Valid {
		payload["purchase_price"] = item.PurchasePrice.Int64
	}
	if item.PriceCurrency.Valid {
		payload["price_currency"] = item.PriceCurrency.String
	}
	if item.StorageLocation.Valid {
		payload["storage_location"] = item.StorageLocation.String
	}
	if item.Season.Valid {
		payload["season"] = item.Season.String
	}
	if item.Color.Valid {
		payload["color"] = item.Color.String
	}
	if item.Size.Valid {
		payload["size"] = item.Size.String
	}

	return payload, nil
}

// snapshotPhoto creates a full photo snapshot.
func (s *PullSnapshotter) snapshotPhoto(ctx context.Context, exec database.Executor, id uuid.UUID) (map[string]any, error) {
	q := sqlc.NewFromExecutor(exec)

	photo, err := q.GetPhoto(ctx, sqlc.GetPhotoParams{
		ID:     id,
		UserID: id, // placeholder
	})
	if err != nil {
		return nil, fmt.Errorf("get photo snapshot: %w", err)
	}

	payload := map[string]any{
		"entity_id": photo.ID,
		"item_id":   photo.ItemID,
		"version":   photo.Version,
		"status":    photo.Status,
		"sort_order": photo.SortOrder,
	}

	if photo.MimeType.Valid {
		payload["mime_type"] = photo.MimeType.String
	}
	if photo.SizeBytes.Valid {
		payload["size_bytes"] = photo.SizeBytes.Int64
	}
	if photo.Width.Valid {
		payload["width"] = photo.Width.Int32
	}
	if photo.Height.Valid {
		payload["height"] = photo.Height.Int32
	}
	if photo.ContentHash.Valid {
		payload["content_hash"] = photo.ContentHash.String
	}

	return payload, nil
}
