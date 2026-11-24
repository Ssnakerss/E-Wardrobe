CREATE TABLE categories (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL
                     REFERENCES users(id) ON DELETE CASCADE,
    parent_id    UUID,
    template_key TEXT,
    name         TEXT NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    version      BIGINT NOT NULL DEFAULT 1,
    archived_at  TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT categories_id_user_unique
        UNIQUE (id, user_id),

    CONSTRAINT categories_name_valid
        CHECK (
            char_length(btrim(name)) BETWEEN 1 AND 200
        ),

    CONSTRAINT categories_sort_order_valid
        CHECK (sort_order >= 0),

    CONSTRAINT categories_version_valid
        CHECK (version > 0),

    CONSTRAINT categories_not_own_parent
        CHECK (
            parent_id IS NULL
            OR parent_id <> id
        ),

    CONSTRAINT categories_parent_fk
        FOREIGN KEY (parent_id, user_id)
        REFERENCES categories(id, user_id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX categories_template_key_unique_idx
    ON categories (user_id, template_key)
    WHERE template_key IS NOT NULL;

CREATE INDEX categories_user_parent_idx
    ON categories (user_id, parent_id, sort_order, name)
    WHERE deleted_at IS NULL;

CREATE TRIGGER categories_set_updated_at
BEFORE UPDATE ON categories
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL
                          REFERENCES users(id) ON DELETE CASCADE,
    category_id       UUID NOT NULL,
    name              TEXT NOT NULL,
    brand             TEXT,
    description       TEXT,
    purchase_year     SMALLINT,
    purchase_price    BIGINT,
    price_currency    TEXT,
    purchase_place    TEXT,
    condition_rating  SMALLINT,
    preference_rating SMALLINT,
    last_worn_on      DATE,
    storage_location  TEXT,
    season            TEXT,
    color             TEXT,
    size              TEXT,
    status            TEXT NOT NULL DEFAULT 'in_use',
    version           BIGINT NOT NULL DEFAULT 1,
    archived_at       TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT items_id_user_unique
        UNIQUE (id, user_id),

    CONSTRAINT items_category_fk
        FOREIGN KEY (category_id, user_id)
        REFERENCES categories(id, user_id)
        ON DELETE RESTRICT,

    CONSTRAINT items_name_valid
        CHECK (
            char_length(btrim(name)) BETWEEN 1 AND 200
        ),

    CONSTRAINT items_purchase_year_valid
        CHECK (
            purchase_year IS NULL
            OR purchase_year BETWEEN 1800 AND 9999
        ),

    CONSTRAINT items_price_nonnegative
        CHECK (
            purchase_price IS NULL
            OR purchase_price >= 0
        ),

    CONSTRAINT items_price_currency_pair
        CHECK (
            (purchase_price IS NULL AND price_currency IS NULL)
            OR
            (purchase_price IS NOT NULL AND price_currency IS NOT NULL)
        ),

    CONSTRAINT items_currency_format
        CHECK (
            price_currency IS NULL
            OR price_currency ~ '^[A-Z]{3}$'
        ),

    CONSTRAINT items_condition_rating_valid
        CHECK (
            condition_rating IS NULL
            OR condition_rating BETWEEN 1 AND 5
        ),

    CONSTRAINT items_preference_rating_valid
        CHECK (
            preference_rating IS NULL
            OR preference_rating BETWEEN 1 AND 5
        ),

    CONSTRAINT items_season_valid
        CHECK (
            season IS NULL
            OR season IN (
                'winter',
                'spring',
                'summer',
                'autumn',
                'demi_season',
                'all_season'
            )
        ),

    CONSTRAINT items_status_valid
        CHECK (
            status IN ('in_use', 'stored', 'needs_repair')
        ),

    CONSTRAINT items_version_valid
        CHECK (version > 0)
);

CREATE INDEX items_user_list_idx
    ON items (user_id, updated_at DESC, id)
    WHERE deleted_at IS NULL;

CREATE INDEX items_user_category_idx
    ON items (user_id, category_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX items_user_archive_idx
    ON items (user_id, archived_at, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX items_storage_location_idx
    ON items (user_id, storage_location)
    INCLUDE (updated_at)
    WHERE deleted_at IS NULL
      AND storage_location IS NOT NULL;

CREATE TRIGGER items_set_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE item_photos (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id           UUID NOT NULL,
    user_id           UUID NOT NULL
                          REFERENCES users(id) ON DELETE CASCADE,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'awaiting_upload',
    object_key        TEXT,
    thumbnail_key     TEXT,
    mime_type         TEXT,
    size_bytes        BIGINT,
    width             INTEGER,
    height            INTEGER,
    content_hash      TEXT,
    processing_error  TEXT,
    version           BIGINT NOT NULL DEFAULT 1,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT item_photos_item_fk
        FOREIGN KEY (item_id, user_id)
        REFERENCES items(id, user_id)
        ON DELETE CASCADE,

    CONSTRAINT item_photos_sort_order_valid
        CHECK (sort_order >= 0),

    CONSTRAINT item_photos_status_valid
        CHECK (
            status IN (
                'awaiting_upload',
                'processing',
                'ready',
                'failed'
            )
        ),

    CONSTRAINT item_photos_size_valid
        CHECK (
            size_bytes IS NULL
            OR size_bytes >= 0
        ),

    CONSTRAINT item_photos_width_valid
        CHECK (
            width IS NULL
            OR width > 0
        ),

    CONSTRAINT item_photos_height_valid
        CHECK (
            height IS NULL
            OR height > 0
        ),

    CONSTRAINT item_photos_version_valid
        CHECK (version > 0),

    CONSTRAINT item_photos_ready_complete
        CHECK (
            status <> 'ready'
            OR (
                object_key IS NOT NULL
                AND thumbnail_key IS NOT NULL
                AND mime_type IS NOT NULL
                AND size_bytes IS NOT NULL
                AND width IS NOT NULL
                AND height IS NOT NULL
            )
        )
);

CREATE UNIQUE INDEX item_photos_object_key_unique_idx
    ON item_photos (object_key)
    WHERE object_key IS NOT NULL;

CREATE UNIQUE INDEX item_photos_thumbnail_key_unique_idx
    ON item_photos (thumbnail_key)
    WHERE thumbnail_key IS NOT NULL;

CREATE INDEX item_photos_item_list_idx
    ON item_photos (item_id, sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX item_photos_pending_upload_idx
    ON item_photos (user_id, created_at)
    WHERE status IN ('awaiting_upload', 'processing', 'failed')
      AND deleted_at IS NULL;

CREATE TRIGGER item_photos_set_updated_at
BEFORE UPDATE ON item_photos
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TABLE IF EXISTS item_photos;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS categories;