-- +goose Up
CREATE TABLE plans (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255)  NOT NULL,
    slug          VARCHAR(255)  NOT NULL UNIQUE,
    description   TEXT          NOT NULL DEFAULT '',
    price         NUMERIC(10,2) NOT NULL DEFAULT 0,
    level         INTEGER       NOT NULL DEFAULT 0,
    is_active     BOOLEAN       NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plans_slug ON plans(slug);
CREATE INDEX idx_plans_level ON plans(level);

CREATE TABLE user_subscriptions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id),
    plan_id    UUID         NOT NULL REFERENCES plans(id),
    status     VARCHAR(20)  NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
CREATE UNIQUE INDEX idx_user_subscriptions_active ON user_subscriptions(user_id) WHERE status = 'active';

ALTER TABLE videos ADD COLUMN minimum_plan_level INTEGER;

-- +goose Down
ALTER TABLE videos DROP COLUMN IF EXISTS minimum_plan_level;
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS plans;
