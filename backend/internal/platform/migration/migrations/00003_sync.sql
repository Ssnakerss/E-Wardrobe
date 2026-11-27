-- +goose Up
CREATE TABLE user_sync_state (
    user_id    UUID PRIMARY KEY
                   REFERENCES users(id) ON DELETE CASCADE,
    last_seq   BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT user_sync_state_last_seq_valid
        CHECK (last_seq >= 0)
);

CREATE TRIGGER user_sync_state_set_updated_at
BEFORE UPDATE ON user_sync_state
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE sync_changes (
    user_id        UUID NOT NULL
                       REFERENCES users(id) ON DELETE CASCADE,
    seq            BIGINT NOT NULL,
    entity         TEXT NOT NULL,
    entity_id      UUID NOT NULL,
    operation      TEXT NOT NULL,
    entity_version BIGINT NOT NULL,
    payload        JSONB NOT NULL,
    changed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT sync_changes_primary_key
        PRIMARY KEY (user_id, seq),

    CONSTRAINT sync_changes_seq_valid
        CHECK (seq > 0),

    CONSTRAINT sync_changes_entity_valid
        CHECK (
            entity IN ('category', 'item', 'photo')
        ),

    CONSTRAINT sync_changes_operation_valid
        CHECK (
            operation IN ('upsert', 'delete')
        ),

    CONSTRAINT sync_changes_version_valid
        CHECK (entity_version > 0),

    CONSTRAINT sync_changes_payload_object
        CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX sync_changes_entity_history_idx
    ON sync_changes (
        user_id,
        entity,
        entity_id,
        seq DESC
    );

CREATE INDEX sync_changes_changed_at_idx
    ON sync_changes (changed_at);

CREATE TABLE processed_mutations (
    user_id      UUID NOT NULL
                     REFERENCES users(id) ON DELETE CASCADE,
    mutation_id  UUID NOT NULL,
    request_hash TEXT NOT NULL,
    entity       TEXT NOT NULL,
    entity_id    UUID NOT NULL,
    result       JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT processed_mutations_primary_key
        PRIMARY KEY (user_id, mutation_id),

    CONSTRAINT processed_mutations_entity_valid
        CHECK (
            entity IN ('category', 'item', 'photo')
        ),

    CONSTRAINT processed_mutations_request_hash_valid
        CHECK (char_length(request_hash) >= 32),

    CONSTRAINT processed_mutations_result_object
        CHECK (jsonb_typeof(result) = 'object')
);

CREATE INDEX processed_mutations_entity_idx
    ON processed_mutations (
        user_id,
        entity,
        entity_id,
        created_at
    );

CREATE INDEX processed_mutations_created_at_idx
    ON processed_mutations (created_at);

-- Retention note:
-- sync_changes and processed_mutations are intentionally not automatically
-- purged in the MVP. Device-aware compaction must be designed before any
-- deletion of synchronization history.

-- +goose Down

DROP TABLE IF EXISTS processed_mutations;
DROP TABLE IF EXISTS sync_changes;
DROP TABLE IF EXISTS user_sync_state;