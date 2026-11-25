-- name: CreatePhoto :one
INSERT INTO item_photos (
    id, item_id, user_id, sort_order, status, object_key, thumbnail_key,
    mime_type, size_bytes, width, height, version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetPhoto :one
SELECT * FROM item_photos
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: GetPhotoByItem :many
SELECT * FROM item_photos
WHERE item_id = $1 AND user_id = $2 AND deleted_at IS NULL
ORDER BY sort_order, created_at;

-- name: UpdatePhoto :one
UPDATE item_photos
SET
    status = COALESCE($3, status),
    object_key = COALESCE($4, object_key),
    thumbnail_key = COALESCE($5, thumbnail_key),
    mime_type = COALESCE($6, mime_type),
    size_bytes = COALESCE($7, size_bytes),
    width = COALESCE($8, width),
    height = COALESCE($9, height),
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $10
RETURNING *;

-- name: SoftDeletePhoto :one
UPDATE item_photos
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: UpdatePhotoSortOrder :one
UPDATE item_photos
SET sort_order = $3, version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $4
RETURNING *;
