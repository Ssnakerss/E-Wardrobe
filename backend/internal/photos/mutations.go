package photos

import (
	"context"
	"fmt"

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

	itemID, _ := payload["item_id"].(uuid.UUID)
	if itemID == uuid.Nil {
		return fmt.Errorf("missing item_id")
	}

	sortOrder, ok := payload["sort_order"].(float64)
	var sortOrderInt32 int32
	if ok {
		sortOrderInt32 = int32(sortOrder)
	}

	input := PhotoInput{
		ItemID:    itemID,
		SortOrder: sortOrderInt32,
	}

	_, err := svc.Create(ctx, tx, userID, itemID, input)
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

	var patch PhotoPatch
	patch.ID = entityID
	patch.BaseVersion = baseVersion

	if status, ok := payload["status"].(string); ok {
		s := PhotoStatus(status)
		patch.Status = &s
	}
	if mimeType, ok := payload["mime_type"].(string); ok {
		patch.MimeType = &mimeType
	}
	if sizeBytes, ok := payload["size_bytes"].(float64); ok {
		v := int64(sizeBytes)
		patch.SizeBytes = &v
	}
	if width, ok := payload["width"].(float64); ok {
		v := int32(width)
		patch.Width = &v
	}
	if height, ok := payload["height"].(float64); ok {
		v := int32(height)
		patch.Height = &v
	}
	if contentHash, ok := payload["content_hash"].(string); ok {
		patch.ContentHash = &contentHash
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
		return fmt.Errorf("photo not found or not owned: %w", err)
	}

	// Get current version
	photo, err := svc.GetByID(ctx, tx, mutation.EntityID, userID)
	if err != nil {
		return fmt.Errorf("get photo: %w", err)
	}

	err = svc.SoftDelete(ctx, tx, mutation.EntityID, userID, photo.Version)
	if err != nil {
		return fmt.Errorf("apply delete: %w", err)
	}

	return nil
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
