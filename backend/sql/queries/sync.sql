-- name: UpsertUserSyncState :one
INSERT INTO user_sync_state (user_id, last_seq)
VALUES ($1, 0)
ON CONFLICT (user_id) DO UPDATE SET last_seq = 0
RETURNING *;

-- name: GetOrCreateUserSyncState :one
SELECT * FROM user_sync_state
WHERE user_id = $1
FOR UPDATE;

-- name: NextSeq :one
UPDATE user_sync_state
SET last_seq = last_seq + 1
WHERE user_id = $1
RETURNING last_seq AS seq;

-- name: CreateSyncChange :_, as
INSERT INTO sync_changes (
    user_id, seq, entity, entity_id, operation, entity_version, payload, changed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, now()
);

-- name: ListSyncChanges :many
SELECT * FROM sync_changes
WHERE user_id = $1 AND seq > $2
ORDER BY seq ASC
LIMIT $3 + 1;

-- name: CreateProcessedMutation :_, as
INSERT INTO processed_mutations (
    user_id, mutation_id, request_hash, entity, entity_id, result
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetProcessedMutation :one
SELECT * FROM processed_mutations
WHERE user_id = $1 AND mutation_id = $2;
