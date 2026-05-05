-- +goose Up
ALTER TABLE users ADD COLUMN stripe_customer_id VARCHAR(255);
CREATE UNIQUE INDEX idx_users_stripe_customer_id ON users(stripe_customer_id)
    WHERE stripe_customer_id IS NOT NULL;

ALTER TABLE plans ADD COLUMN stripe_price_id VARCHAR(255);
CREATE INDEX idx_plans_stripe_price_id ON plans(stripe_price_id)
    WHERE stripe_price_id IS NOT NULL;

ALTER TABLE user_subscriptions ADD COLUMN stripe_subscription_id VARCHAR(255);
ALTER TABLE user_subscriptions ADD COLUMN stripe_customer_id VARCHAR(255);
CREATE UNIQUE INDEX idx_user_subscriptions_stripe_subscription_id
    ON user_subscriptions(stripe_subscription_id)
    WHERE stripe_subscription_id IS NOT NULL;
CREATE INDEX idx_user_subscriptions_stripe_customer_id
    ON user_subscriptions(stripe_customer_id)
    WHERE stripe_customer_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_user_subscriptions_stripe_customer_id;
DROP INDEX IF EXISTS idx_user_subscriptions_stripe_subscription_id;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS stripe_subscription_id;

DROP INDEX IF EXISTS idx_plans_stripe_price_id;
ALTER TABLE plans DROP COLUMN IF EXISTS stripe_price_id;

DROP INDEX IF EXISTS idx_users_stripe_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
