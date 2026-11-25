-- name: UpsertCategory :one
INSERT INTO categories (
    id, user_id, parent_id, template_key, name, sort_order, version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id) DO UPDATE
SET
    parent_id = EXCLUDED.parent_id,
    template_key = EXCLUDED.template_key,
    name = EXCLUDED.name,
    sort_order = EXCLUDED.sort_order,
    version = EXCLUDED.version,
    updated_at = now()
RETURNING *;

-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: ListCategories :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY sort_order, name;

-- name: UpdateCategory :one
UPDATE categories
SET
    parent_id = COALESCE($3, parent_id),
    template_key = COALESCE($4, template_key),
    name = COALESCE($5, name),
    sort_order = COALESCE($6, sort_order),
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $7
RETURNING *;

-- name: SoftDeleteCategory :one
UPDATE categories
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: CheckCategoryCycle :bool
WITH RECURSIVE category_tree AS (
    SELECT id, parent_id FROM categories WHERE id = $1
    UNION ALL
    SELECT c.id, c.parent_id FROM categories c
    INNER JOIN category_tree ct ON c.id = ct.parent_id
)
SELECT EXISTS (
    SELECT 1 FROM category_tree WHERE id = $2
);

-- name: CountItemsInCategory :one
SELECT COUNT(*) FROM items
WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CountChildCategories :one
SELECT COUNT(*) FROM categories
WHERE parent_id = $1 AND user_id = $2 AND deleted_at IS NULL;
