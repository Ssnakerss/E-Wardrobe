-- name: CreateItem :one
INSERT INTO items (
    id, user_id, category_id, name, brand, description,
    purchase_year, purchase_price, price_currency, purchase_place,
    condition_rating, preference_rating, last_worn_on,
    storage_location, season, color, size, status, version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
)
RETURNING *;

-- name: GetItem :one
SELECT * FROM items
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateItem :one
UPDATE items
SET
    category_id = COALESCE($3, category_id),
    name = COALESCE($4, name),
    brand = $5,
    description = $6,
    purchase_year = $7,
    purchase_price = $8,
    price_currency = $9,
    purchase_place = $10,
    condition_rating = $11,
    preference_rating = $12,
    last_worn_on = $13,
    storage_location = $14,
    season = $15,
    color = $16,
    size = $17,
    status = COALESCE($18, status),
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $19
RETURNING *;

-- name: SoftDeleteItem :one
UPDATE items
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: ArchiveItem :one
UPDATE items
SET archived_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: UnarchiveItem :one
UPDATE items
SET archived_at = NULL, version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: ListItems :many
SELECT * FROM items
WHERE user_id = $1
  AND deleted_at IS NULL
  AND ($7::boolean IS FALSE OR archived_at IS NULL)
  AND ($2::uuid IS NULL OR category_id = $2 OR category_id IN (
      SELECT id FROM categories WHERE user_id = $1 AND deleted_at IS NULL
  ))
  AND ($3::text IS NULL OR
       LOWER(name) LIKE LOWER($3) OR
       LOWER(brand) LIKE LOWER($3) OR
       LOWER(description) LIKE LOWER($3))
  AND ($4::text IS NULL OR status = $4)
  AND ($5::text IS NULL OR season = $5)
  AND ($6::smallint IS NULL OR condition_rating = $6)
ORDER BY updated_at DESC, id DESC
LIMIT $8 OFFSET $9;

-- name: CountItems :one
SELECT COUNT(*) FROM items
WHERE user_id = $1
  AND deleted_at IS NULL
  AND ($5::boolean IS FALSE OR archived_at IS NULL)
  AND ($2::uuid IS NULL OR category_id = $2)
  AND ($3::text IS NULL OR
       LOWER(name) LIKE LOWER($3) OR
       LOWER(brand) LIKE LOWER($3) OR
       LOWER(description) LIKE LOWER($3))
  AND ($4::text IS NULL OR status = $4);

-- name: ListStorageLocationSuggestions :many
SELECT DISTINCT storage_location
FROM items
WHERE user_id = $1
  AND storage_location IS NOT NULL
  AND deleted_at IS NULL
  AND LOWER(storage_location) LIKE LOWER('%' || $2 || '%')
ORDER BY storage_location
LIMIT $3;
