package categories

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// Service handles business logic for categories.
type Service struct {
	repo    *Repository
	journal *changejournal.Writer
}

// NewService creates a new categories service.
func NewService(repo *Repository, journal *changejournal.Writer) *Service {
	return &Service{repo: repo, journal: journal}
}

// Create creates a new category.
func (s *Service) Create(ctx context.Context, tx database.Tx, userID uuid.UUID, input CategoryInput) (Category, error) {
	cat := Category{
		ID:        uuid.New(),
		UserID:    userID,
		ParentID:  input.ParentID,
		Name:      input.Name,
		SortOrder: input.SortOrder,
		Version:   1,
	}

	cat, err := s.repo.Create(ctx, tx, cat)
	if err != nil {
		return cat, fmt.Errorf("create category: %w", err)
	}

	// Write to change journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "category",
		EntityID:      cat.ID,
		Operation:     "upsert",
		EntityVersion: cat.Version,
		Payload:       map[string]any{"name": cat.Name, "parent_id": cat.ParentID},
	})
	if err != nil {
		return cat, fmt.Errorf("write journal: %w", err)
	}

	return cat, nil
}

// GetByID returns a category by ID.
func (s *Service) GetByID(ctx context.Context, tx database.Tx, id, userID uuid.UUID) (Category, error) {
	return s.repo.GetByID(ctx, tx, id, userID)
}

// List returns all categories for a user.
func (s *Service) List(ctx context.Context, tx database.Tx, userID uuid.UUID) ([]Category, error) {
	return s.repo.List(ctx, tx, userID)
}

// Update updates a category.
func (s *Service) Update(ctx context.Context, tx database.Tx, userID uuid.UUID, patch CategoryPatch) (Category, error) {
	// Verify ownership
	existing, err := s.repo.GetByID(ctx, tx, patch.ID, userID)
	if err != nil {
		return existing, fmt.Errorf("category not found or not owned: %w", err)
	}

	// Check cycle if parent is changing
	if patch.ParentID != nil && existing.ParentID != nil && *patch.ParentID != *existing.ParentID {
		isCycle, err := s.repo.CheckCycle(ctx, tx, patch.ID, *patch.ParentID)
		if err != nil {
			return existing, fmt.Errorf("check cycle: %w", err)
		}
		if isCycle {
			return existing, fmt.Errorf("moving category would create a cycle")
		}
	}

	cat, err := s.repo.Update(ctx, tx, patch)
	if err != nil {
		return cat, fmt.Errorf("update category: %w", err)
	}

	// Write to change journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "category",
		EntityID:      cat.ID,
		Operation:     "upsert",
		EntityVersion: cat.Version,
		Payload:       map[string]any{"name": cat.Name, "parent_id": cat.ParentID},
	})
	if err != nil {
		return cat, fmt.Errorf("write journal: %w", err)
	}

	return cat, nil
}

// SoftDelete deletes a category.
func (s *Service) SoftDelete(ctx context.Context, tx database.Tx, id, userID uuid.UUID, version int64) error {
	// Verify ownership
	_, err := s.repo.GetByID(ctx, tx, id, userID)
	if err != nil {
		return fmt.Errorf("category not found or not owned: %w", err)
	}

	// Check if category is in use
	itemsCount, err := s.repo.CountItemsInCategory(ctx, tx, id, userID)
	if err != nil {
		return fmt.Errorf("count items: %w", err)
	}
	if itemsCount > 0 {
		return fmt.Errorf("cannot delete category with active items")
	}

	// Check for child categories
	childrenCount, err := s.repo.CountChildCategories(ctx, tx, id, userID)
	if err != nil {
		return fmt.Errorf("count children: %w", err)
	}
	if childrenCount > 0 {
		return fmt.Errorf("cannot delete category with child categories")
	}

	// Soft delete
	err = s.repo.SoftDelete(ctx, tx, id, userID, version)
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}

	// Write tombstone to journal
	err = s.journal.WriteRecord(ctx, tx, changejournal.Change{
		UserID:        userID,
		Entity:        "category",
		EntityID:      id,
		Operation:     "delete",
		EntityVersion: version + 1,
		Payload:       map[string]any{"tombstone": true},
	})
	if err != nil {
		return fmt.Errorf("write tombstone: %w", err)
	}

	return nil
}
