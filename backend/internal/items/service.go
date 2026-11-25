package items

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/categories"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// Service handles business logic for items.
type Service struct {
	repo    *Repository
	catRepo *categories.Repository
	journal *changejournal.Writer
}

// NewService creates a new items service.
func NewService(repo *Repository, catRepo *categories.Repository, journal *changejournal.Writer) *Service {
	return &Service{repo: repo, catRepo: catRepo, journal: journal}
}

// Create validates and creates a new item.
func (s *Service) Create(ctx context.Context, exec database.Executor, userID uuid.UUID, input ItemInput) (Item, error) {
	_, err := s.catRepo.GetByID(ctx, exec, input.CategoryID, userID)
	if err != nil {
		return Item{}, fmt.Errorf("invalid category: %w", err)
	}

	if err := s.validateCreateInput(input); err != nil {
		return Item{}, err
	}

	if input.StorageLocation != nil {
		n := normalizeStorageLocation(*input.StorageLocation)
		if n == "" {
			input.StorageLocation = nil
		} else {
			input.StorageLocation = &n
		}
	}

	item := Item{
		ID:               uuid.New(),
		UserID:           userID,
		CategoryID:       input.CategoryID,
		Name:             strings.TrimSpace(input.Name),
		Brand:            input.Brand,
		Description:      input.Description,
		PurchaseYear:     input.PurchaseYear,
		PurchasePrice:    input.PurchasePrice,
		PriceCurrency:    input.PriceCurrency,
		PurchasePlace:    input.PurchasePlace,
		ConditionRating:  input.ConditionRating,
		PreferenceRating: input.PreferenceRating,
		LastWornOn:       input.LastWornOn,
		StorageLocation:  input.StorageLocation,
		Season:           input.Season,
		Color:            input.Color,
		Size:             input.Size,
		Status:           input.Status,
		Version:          1,
	}

	item, err = s.repo.Create(ctx, exec, item)
	if err != nil {
		return item, fmt.Errorf("create item: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
		UserID:        userID,
		Entity:        "item",
		EntityID:      item.ID,
		Operation:     "upsert",
		EntityVersion: item.Version,
		Payload:       s.itemToPayload(item),
	})
	if err != nil {
		return item, fmt.Errorf("write journal: %w", err)
	}

	return item, nil
}

// GetByID returns an item by ID with ownership check.
func (s *Service) GetByID(ctx context.Context, exec database.Executor, id, userID uuid.UUID) (Item, error) {
	return s.repo.GetByID(ctx, exec, id, userID)
}

// Update validates and updates an item.
func (s *Service) Update(ctx context.Context, exec database.Executor, userID uuid.UUID, patch ItemPatch) (Item, error) {
	existing, err := s.repo.GetByID(ctx, exec, patch.ID, userID)
	if err != nil {
		return Item{}, fmt.Errorf("item not found or not owned: %w", err)
	}

	if patch.BaseVersion != 0 && patch.BaseVersion != existing.Version {
		return Item{}, fmt.Errorf("version conflict: expected %d, got %d", existing.Version, patch.BaseVersion)
	}

	if patch.CategoryID != nil && *patch.CategoryID != existing.CategoryID {
		_, err := s.catRepo.GetByID(ctx, exec, *patch.CategoryID, userID)
		if err != nil {
			return Item{}, fmt.Errorf("invalid category: %w", err)
		}
	}

	if err := s.validatePartialInput(existing, patch); err != nil {
		return Item{}, err
	}

	if patch.StorageLocation != nil {
		n := normalizeStorageLocation(*patch.StorageLocation)
		if n == "" {
			patch.StorageLocation = nil
		} else {
			patch.StorageLocation = &n
		}
	}

	updated, err := s.repo.Update(ctx, exec, patch)
	if err != nil {
		return updated, fmt.Errorf("update item: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
		UserID:        userID,
		Entity:        "item",
		EntityID:      updated.ID,
		Operation:     "upsert",
		EntityVersion: updated.Version,
		Payload:       s.itemToPayload(updated),
	})
	if err != nil {
		return updated, fmt.Errorf("write journal: %w", err)
	}

	return updated, nil
}

// SoftDelete marks an item as deleted.
func (s *Service) SoftDelete(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) error {
	_, err := s.repo.GetByID(ctx, exec, id, userID)
	if err != nil {
		return fmt.Errorf("item not found or not owned: %w", err)
	}

	err = s.repo.SoftDelete(ctx, exec, id, userID, version)
	if err != nil {
		return fmt.Errorf("soft delete item: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
		UserID:        userID,
		Entity:        "item",
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

// Archive marks an item as archived.
func (s *Service) Archive(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) (Item, error) {
	_, err := s.repo.GetByID(ctx, exec, id, userID)
	if err != nil {
		return Item{}, fmt.Errorf("item not found or not owned: %w", err)
	}

	item, err := s.repo.Archive(ctx, exec, id, userID, version)
	if err != nil {
		return item, fmt.Errorf("archive item: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
		UserID:        userID,
		Entity:        "item",
		EntityID:      item.ID,
		Operation:     "upsert",
		EntityVersion: item.Version,
		Payload:       s.itemToPayload(item),
	})
	if err != nil {
		return item, fmt.Errorf("write journal: %w", err)
	}

	return item, nil
}

// Unarchive removes the archive flag from an item.
func (s *Service) Unarchive(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) (Item, error) {
	_, err := s.repo.GetByID(ctx, exec, id, userID)
	if err != nil {
		return Item{}, fmt.Errorf("item not found or not owned: %w", err)
	}

	item, err := s.repo.Unarchive(ctx, exec, id, userID, version)
	if err != nil {
		return item, fmt.Errorf("unarchive item: %w", err)
	}

	err = s.journal.WriteRecord(ctx, exec, changejournal.Change{
		UserID:        userID,
		Entity:        "item",
		EntityID:      item.ID,
		Operation:     "upsert",
		EntityVersion: item.Version,
		Payload:       s.itemToPayload(item),
	})
	if err != nil {
		return item, fmt.Errorf("write journal: %w", err)
	}

	return item, nil
}

// List returns items matching the options.
func (s *Service) List(ctx context.Context, exec database.Executor, opts ListOptions) ([]Item, error) {
	return s.repo.List(ctx, exec, opts)
}

// CountItems returns the count of items matching the options.
func (s *Service) CountItems(ctx context.Context, exec database.Executor, opts ListOptions) (int64, error) {
	return s.repo.CountItems(ctx, exec, opts)
}

// ListStorageLocationSuggestions returns distinct storage locations matching the query.
func (s *Service) ListStorageLocationSuggestions(ctx context.Context, exec database.Executor, userID uuid.UUID, query string, limit int32) ([]string, error) {
	return s.repo.ListStorageLocationSuggestions(ctx, exec, userID, query, limit)
}

// validateCreateInput validates the input for creating an item.
func (s *Service) validateCreateInput(input ItemInput) *AppError {
	if (input.PurchasePrice != nil) != (input.PriceCurrency != nil) {
		return ValidationError("price and currency must be provided together")
	}
	if input.ConditionRating != nil {
		if *input.ConditionRating < 1 || *input.ConditionRating > 5 {
			return ValidationError("condition rating must be between 1 and 5")
		}
	}
	if input.PreferenceRating != nil {
		if *input.PreferenceRating < 1 || *input.PreferenceRating > 5 {
			return ValidationError("preference rating must be between 1 and 5")
		}
	}
	switch input.Status {
	case "in_use", "stored", "needs_repair":
	default:
		return ValidationError("invalid status: must be in_use, stored, or needs_repair")
	}
	if input.Season != nil {
		switch *input.Season {
		case "winter", "spring", "summer", "autumn", "demi_season", "all_season", "":
		default:
			return ValidationError("invalid season")
		}
	}
	return nil
}

// validatePartialInput validates fields in a patch.
func (s *Service) validatePartialInput(existing Item, patch ItemPatch) *AppError {
	if (patch.PurchasePrice != nil) != (patch.PriceCurrency != nil) {
		return ValidationError("price and currency must be provided together")
	}
	if patch.ConditionRating != nil {
		if *patch.ConditionRating < 1 || *patch.ConditionRating > 5 {
			return ValidationError("condition rating must be between 1 and 5")
		}
	}
	if patch.PreferenceRating != nil {
		if *patch.PreferenceRating < 1 || *patch.PreferenceRating > 5 {
			return ValidationError("preference rating must be between 1 and 5")
		}
	}
	if patch.Status != nil {
		switch *patch.Status {
		case "in_use", "stored", "needs_repair", "":
		default:
			return ValidationError("invalid status: must be in_use, stored, or needs_repair")
		}
	}
	if patch.Season != nil {
		switch *patch.Season {
		case "winter", "spring", "summer", "autumn", "demi_season", "all_season", "":
		default:
			return ValidationError("invalid season")
		}
	}
	return nil
}

// itemToPayload converts an item to a journal payload map.
func (s *Service) itemToPayload(item Item) map[string]any {
	return map[string]any{
		"name":               item.Name,
		"category_id":        item.CategoryID,
		"brand":              item.Brand,
		"description":        item.Description,
		"purchase_year":      item.PurchaseYear,
		"purchase_price":     item.PurchasePrice,
		"price_currency":     item.PriceCurrency,
		"purchase_place":     item.PurchasePlace,
		"condition_rating":   item.ConditionRating,
		"preference_rating":  item.PreferenceRating,
		"last_worn_on":       item.LastWornOn,
		"storage_location":   item.StorageLocation,
		"season":             item.Season,
		"color":              item.Color,
		"size":               item.Size,
		"status":             item.Status,
	}
}

// normalizeStorageLocation trims whitespace and collapses multiple spaces.
func normalizeStorageLocation(s string) string {
	trimmed := strings.TrimSpace(s)
	var result []rune
	prevSpace := false
	for _, r := range trimmed {
		if r == ' ' {
			if !prevSpace {
				result = append(result, r)
			}
			prevSpace = true
		} else {
			prevSpace = false
			result = append(result, r)
		}
	}
	return string(result)
}
