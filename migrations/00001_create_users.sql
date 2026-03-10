-- +goose Up
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255)  NOT NULL UNIQUE,
    password_hash VARCHAR(255)  NOT NULL,
    role          VARCHAR(20)   NOT NULL DEFAULT 'user',
    status        VARCHAR(20)   NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE user_profiles (
    user_id      UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(255)  NOT NULL DEFAULT '',
    avatar_url   VARCHAR(1024) NOT NULL DEFAULT '',
    date_of_birth DATE,
    language     VARCHAR(5)    NOT NULL DEFAULT 'en',
    country      VARCHAR(2)    NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
