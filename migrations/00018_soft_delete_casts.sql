-- +goose Up
CREATE TYPE cast_status AS ENUM ('active', 'archived');

ALTER TABLE casts
    ADD COLUMN status cast_status NOT NULL DEFAULT 'active';

CREATE INDEX idx_casts_status ON casts(status);

-- +goose Down
DROP INDEX IF EXISTS idx_casts_status;
ALTER TABLE casts DROP COLUMN status;
DROP TYPE IF EXISTS cast_status;
