package categories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database/sqlc"
	"github.com/google/uuid"
)

// Repository handles database operations for categories.
type Repository struct {
	pool *database.Pool
}

// NewRepository creates a new categories repository.
func NewRepository(pool *database.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new category.
func (r *Repository) Create(ctx context.Context, exec database.Executor, cat Category) (Category, error) {
	q := sqlc.NewFromExecutor(exec)

	var parentIDSQL *uuid.UUID
	if cat.ParentID != nil {
		parentIDSQL = cat.ParentID
	}

	var templateKeySQL sql.NullString
	if cat.TemplateKey != nil {
		templateKeySQL = sql.NullString{String: *cat.TemplateKey, Valid: true}
	}

	c, err := q.UpsertCategory(ctx, sqlc.UpsertCategoryParams{
		ID:          cat.ID,
		UserID:      cat.UserID,
		ParentID:    parentIDSQL,
		TemplateKey: templateKeySQL,
		Name:        cat.Name,
		SortOrder:   cat.SortOrder,
		Version:     cat.Version,
	})
	if err != nil {
		return cat, fmt.Errorf("create category: %w", err)
	}

	return toCategory(c), nil
}

// GetByID returns a category by ID.
func (r *Repository) GetByID(ctx context.Context, exec database.Executor, id, userID uuid.UUID) (Category, error) {
	q := sqlc.NewFromExecutor(exec)

	c, err := q.GetCategory(ctx, sqlc.GetCategoryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return Category{}, fmt.Errorf("get category: %w", err)
	}

	return toCategory(c), nil
}

// List returns all categories for a user.
func (r *Repository) List(ctx context.Context, exec database.Executor, userID uuid.UUID) ([]Category, error) {
	q := sqlc.NewFromExecutor(exec)

	cats, err := q.ListCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	result := make([]Category, len(cats))
	for i, c := range cats {
		result[i] = toCategory(c)
	}

	return result, nil
}

// Update updates a category.
func (r *Repository) Update(ctx context.Context, exec database.Executor, patch CategoryPatch) (Category, error) {
	q := sqlc.NewFromExecutor(exec)

	var parentIDSQL *uuid.UUID
	if patch.ParentID != nil {
		parentIDSQL = patch.ParentID
	}

	var nameSQL *string
	if patch.Name != nil {
		nameSQL = patch.Name
	}

	var sortOrderSQL *int32
	if patch.SortOrder != nil {
		sortOrderSQL = patch.SortOrder
	}

	var templateKeySQL sql.NullString
	c, err := q.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:          patch.ID,
		UserID:      patch.UserID,
		ParentID:    parentIDSQL,
		TemplateKey: templateKeySQL,
		Name:        nullableStringToSQL(nameSQL),
		SortOrder:   nullableInt32ToSQL(sortOrderSQL),
		Version:     patch.BaseVersion,
	})
	if err != nil {
		return Category{}, fmt.Errorf("update category: %w", err)
	}

	return toCategory(c), nil
}

// SoftDelete marks a category as deleted.
func (r *Repository) SoftDelete(ctx context.Context, exec database.Executor, id, userID uuid.UUID, version int64) error {
	q := sqlc.NewFromExecutor(exec)

	_, err := q.SoftDeleteCategory(ctx, sqlc.SoftDeleteCategoryParams{
		ID:      id,
		UserID:  userID,
		Version: version,
	})
	if err != nil {
		return fmt.Errorf("soft delete category: %w", err)
	}

	return nil
}

// CheckCycle checks if moving a category would create a cycle.
func (r *Repository) CheckCycle(ctx context.Context, exec database.Executor, categoryID, parentID uuid.UUID) (bool, error) {
	q := sqlc.NewFromExecutor(exec)

	return q.CheckCategoryCycle(ctx, sqlc.CheckCategoryCycleParams{
		ID:     categoryID,
		RootID: parentID,
	})
}

// CountItemsInCategory returns the number of active items in a category.
func (r *Repository) CountItemsInCategory(ctx context.Context, exec database.Executor, categoryID, userID uuid.UUID) (int64, error) {
	q := sqlc.NewFromExecutor(exec)

	count, err := q.CountItemsInCategory(ctx, sqlc.CountItemsInCategoryParams{
		CategoryID: categoryID,
		UserID:     userID,
	})
	if err != nil {
		return 0, fmt.Errorf("count items in category: %w", err)
	}

	return count, nil
}

// CountChildCategories returns the number of active child categories.
func (r *Repository) CountChildCategories(ctx context.Context, exec database.Executor, parentID, userID uuid.UUID) (int64, error) {
	q := sqlc.NewFromExecutor(exec)

	count, err := q.CountChildCategories(ctx, sqlc.CountChildCategoriesParams{
		ParentID: parentID,
		UserID:   userID,
	})
	if err != nil {
		return 0, fmt.Errorf("count child categories: %w", err)
	}

	return count, nil
}

func toCategory(c sqlc.Category) Category {
	return Category{
		ID:          c.ID,
		UserID:      c.UserID,
		ParentID:    c.ParentID,
		TemplateKey: nullableStringFromSQL(c.TemplateKey),
		Name:        c.Name,
		SortOrder:   c.SortOrder,
		Version:     c.Version,
		ArchivedAt:  nullableTimeFromSQL(c.ArchivedAt),
		DeletedAt:   nullableTimeFromSQL(c.DeletedAt),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func nullableStringFromSQL(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func nullableTimeFromSQL(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func nullableStringToSQL(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullableInt32ToSQL(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}
