-- +goose Up
CREATE TABLE user_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash          VARCHAR(64)  NOT NULL,
    device_name         VARCHAR(255) NOT NULL DEFAULT '',
    user_agent          VARCHAR(512) NOT NULL DEFAULT '',
    ip_address          VARCHAR(45)  NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_activity_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    idle_expires_at     TIMESTAMPTZ  NOT NULL,
    absolute_expires_at TIMESTAMPTZ  NOT NULL,
    revoked_at          TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX idx_user_sessions_token_hash       ON user_sessions(token_hash);
CREATE        INDEX idx_user_sessions_user_active      ON user_sessions(user_id, revoked_at);
CREATE        INDEX idx_user_sessions_absolute_expires ON user_sessions(absolute_expires_at);

-- +goose Down
DROP TABLE IF EXISTS user_sessions;
