-- +goose Up
ALTER TABLE user_profiles
  ADD COLUMN handle      VARCHAR(50)  UNIQUE,
  ADD COLUMN pronouns    VARCHAR(50),
  ADD COLUMN headline    VARCHAR(200),
  ADD COLUMN bio         TEXT,
  ADD COLUMN links       JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN preferences JSONB NOT NULL DEFAULT '{}',
  ADD COLUMN privacy     JSONB NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS handle,
  DROP COLUMN IF EXISTS pronouns,
  DROP COLUMN IF EXISTS headline,
  DROP COLUMN IF EXISTS bio,
  DROP COLUMN IF EXISTS links,
  DROP COLUMN IF EXISTS preferences,
  DROP COLUMN IF EXISTS privacy;
