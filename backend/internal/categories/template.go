package categories

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// BootstrapTemplate is the default category tree copied on registration.
var BootstrapTemplate = []CategoryInput{
	{Name: "Одежда", SortOrder: 1},
	{Name: "Верхняя одежда", SortOrder: 2},
	{Name: "Футболки", SortOrder: 3},
	{Name: "Брюки", SortOrder: 4},
	{Name: "Нижнее бельё", SortOrder: 5},
	{Name: "Обувь", SortOrder: 6},
	{Name: "Аксессуары", SortOrder: 7},
	{Name: "Прочее", SortOrder: 8},
}

// BootstrapCategories creates the default category tree for a new user.
func (s *Service) BootstrapCategories(ctx context.Context, tx database.Tx, userID uuid.UUID) error {
	// Create root categories first
	roots := make(map[string]uuid.UUID)
	for _, input := range BootstrapTemplate {
		cat, err := s.Create(ctx, tx, userID, input)
		if err != nil {
			return fmt.Errorf("bootstrap category %q: %w", input.Name, err)
		}
		roots[input.Name] = cat.ID
	}

	return nil
}
