-- +goose Up
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    tier VARCHAR(20) NOT NULL,
    price_amount BIGINT NOT NULL,
    price_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_interval VARCHAR(20) NOT NULL DEFAULT 'monthly',
    max_quality VARCHAR(10) NOT NULL DEFAULT 'HD',
    max_streams INT NOT NULL DEFAULT 1,
    downloads_allowed BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    external_id VARCHAR(255) NOT NULL DEFAULT '',
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    canceled_at TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

-- Seed default plans
INSERT INTO plans (id, name, tier, price_amount, price_currency, billing_interval, max_quality, max_streams, downloads_allowed) VALUES
    (uuid_generate_v4(), 'Basic', 'basic', 799, 'USD', 'monthly', 'SD', 1, FALSE),
    (uuid_generate_v4(), 'Standard', 'standard', 1299, 'USD', 'monthly', 'FHD', 2, TRUE),
    (uuid_generate_v4(), 'Premium', 'premium', 1799, 'USD', 'monthly', 'UHD', 4, TRUE);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
