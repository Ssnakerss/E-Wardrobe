-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_email_normalized
        CHECK (email = lower(btrim(email)) AND length(email) > 3),

    CONSTRAINT users_status_valid
        CHECK (status IN ('active', 'disabled'))
);

CREATE UNIQUE INDEX users_email_unique_idx
    ON users (email);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash      TEXT NOT NULL UNIQUE,
    email           TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    used_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    revoked_at      TIMESTAMPTZ,
    created_by      TEXT NOT NULL DEFAULT 'cli',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT invitations_email_normalized
        CHECK (
            email IS NULL
            OR email = lower(btrim(email))
        ),

    CONSTRAINT invitations_usage_consistent
        CHECK (
            (used_at IS NULL AND used_by_user_id IS NULL)
            OR
            (used_at IS NOT NULL AND used_by_user_id IS NOT NULL)
        ),

    CONSTRAINT invitations_expiration_valid
        CHECK (expires_at > created_at)
);

CREATE INDEX invitations_email_idx
    ON invitations (email)
    WHERE email IS NOT NULL;

CREATE INDEX invitations_active_idx
    ON invitations (expires_at)
    WHERE used_at IS NULL AND revoked_at IS NULL;

CREATE TABLE sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL
                         REFERENCES users(id) ON DELETE CASCADE,
    token_hash       TEXT NOT NULL UNIQUE,
    csrf_token_hash  TEXT,
    client           TEXT NOT NULL,
    device_name      TEXT,
    expires_at       TIMESTAMPTZ NOT NULL,
    last_used_at     TIMESTAMPTZ,
    revoked_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT sessions_client_valid
        CHECK (client IN ('web', 'mobile')),

    CONSTRAINT sessions_expiration_valid
        CHECK (expires_at > created_at),

    CONSTRAINT sessions_csrf_for_web
        CHECK (
            client <> 'web'
            OR csrf_token_hash IS NOT NULL
        )
);

CREATE INDEX sessions_user_active_idx
    ON sessions (user_id, expires_at)
    WHERE revoked_at IS NULL;

CREATE INDEX sessions_expiration_idx
    ON sessions (expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE account_recovery_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL
                   REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT recovery_expiration_valid
        CHECK (expires_at > created_at)
);

CREATE INDEX recovery_tokens_user_idx
    ON account_recovery_tokens (user_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS account_recovery_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();

-- pgcrypto intentionally remains installed:
-- other database objects or applications may also depend on it.