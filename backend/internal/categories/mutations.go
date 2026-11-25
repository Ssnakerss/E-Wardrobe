package categories

import (
	"context"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// ApplyCreate applies a create mutation from the sync protocol.
func (s *Service) ApplyCreate(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change) error {
	// Parse payload and create category
	// This is a simplified version - in production you'd parse the payload properly
	payload := mutation.Payload

	name, _ := payload["name"].(string)
	parentID, _ := payload["parent_id"].(string)

	var parentUUID *uuid.UUID
	if parentID != "" {
		p, err := uuid.Parse(parentID)
		if err == nil {
			parentUUID = &p
		}
	}

	cat := Category{
		ID:        mutation.EntityID,
		UserID:    userID,
		ParentID:  parentUUID,
		Name:      name,
		SortOrder: 0,
		Version:   mutation.EntityVersion,
	}

	_, err := s.repo.Create(ctx, tx, cat)
	if err != nil {
		return err
	}

	return nil
}

// ApplyUpdate applies an update mutation from the sync protocol.
func (s *Service) ApplyUpdate(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change) error {
	// Parse payload and update category
	payload := mutation.Payload

	name, _ := payload["name"].(string)
	parentID, _ := payload["parent_id"].(string)

	var parentUUID *uuid.UUID
	if parentID != "" {
		p, err := uuid.Parse(parentID)
		if err == nil {
			parentUUID = &p
		}
	}

	patch := CategoryPatch{
		ID:          mutation.EntityID,
		UserID:      userID,
		BaseVersion: mutation.EntityVersion - 1,
		ParentID:    parentUUID,
		Name:        &name,
	}

	_, err := s.repo.Update(ctx, tx, patch)
	if err != nil {
		return err
	}

	return nil
}

// ApplyDelete applies a delete mutation from the sync protocol.
func (s *Service) ApplyDelete(ctx context.Context, tx database.Tx, userID uuid.UUID, mutation changejournal.Change) error {
	// Get current version
	existing, err := s.repo.GetByID(ctx, tx, mutation.EntityID, userID)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, tx, mutation.EntityID, userID, existing.Version)
}
