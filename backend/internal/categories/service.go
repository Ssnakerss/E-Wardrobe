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
func (s *Service) Create(ctx context.Context, exec database.Executor, userID uuid.UUID, input CategoryInput) (Category, error) {
	cat := Category{
		ID:        uuid.New(),
		UserID:    userID,
		ParentID:  input.ParentID,
		Name:      input.Name,
		SortOrder: input.SortOrder,
		Version:   1,
	}

	cat, err := s.repo.Create(ctx, exec, cat)
	if err != nil {
		return cat, fmt.Errorf("create category: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
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
func (s *Service) GetByID(ctx context.Context, exec database.Executor, id, userID uuid.UUID) (Category, error) {
	return s.repo.GetByID(ctx, exec, id, userID)
}

// List returns all categories for a user.
func (s *Service) List(ctx context.Context, exec database.Executor, userID uuid.UUID) ([]Category, error) {
	return s.repo.List(ctx, exec, userID)
}

// Update updates a category.
func (s *Service) Update(ctx context.Context, exec database.Executor, userID uuid.UUID, patch CategoryPatch) (Category, error) {
	existing, err := s.repo.GetByID(ctx, exec, patch.ID, userID)
	if err != nil {
		return existing, fmt.Errorf("category not found or not owned: %w", err)
	}

	if patch.ParentID != nil && existing.ParentID != nil && *patch.ParentID != *existing.ParentID {
		isCycle, err := s.repo.CheckCycle(ctx, exec, patch.ID, *patch.ParentID)
		if err != nil {
			return existing, fmt.Errorf("check cycle: %w", err)
		}
		if isCycle {
			return existing, fmt.Errorf("moving category would create a cycle")
		}
	}

	cat, err := s.repo.Update(ctx, exec, patch)
	if err != nil {
		return cat, fmt.Errorf("update category: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
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
func (s *Service) SoftDelete(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) error {
	_, err := s.repo.GetByID(ctx, exec, id, userID)
	if err != nil {
		return fmt.Errorf("category not found or not owned: %w", err)
	}

	itemsCount, err := s.repo.CountItemsInCategory(ctx, exec, id, userID)
	if err != nil {
		return fmt.Errorf("count items: %w", err)
	}
	if itemsCount > 0 {
		return fmt.Errorf("cannot delete category with active items")
	}

	childrenCount, err := s.repo.CountChildCategories(ctx, exec, id, userID)
	if err != nil {
		return fmt.Errorf("count children: %w", err)
	}
	if childrenCount > 0 {
		return fmt.Errorf("cannot delete category with child categories")
	}

	err = s.repo.SoftDelete(ctx, exec, id, userID, version)
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
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
