-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE video_status ADD VALUE IF NOT EXISTS 'failed';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS transcode_error TEXT NOT NULL DEFAULT '';

-- +goose Down
-- +goose StatementBegin
ALTER TABLE videos DROP COLUMN IF EXISTS transcode_error;
-- Postgres cannot drop a value from an enum without rebuilding the type;
-- reset any rows still on 'failed' to 'draft' so the value is unused, then rebuild.
UPDATE videos SET status = 'draft' WHERE status = 'failed';
ALTER TABLE videos ALTER COLUMN status DROP DEFAULT;
ALTER TABLE videos ALTER COLUMN status TYPE TEXT;
DROP TYPE video_status;
CREATE TYPE video_status AS ENUM ('draft', 'processing', 'published', 'archived');
ALTER TABLE videos ALTER COLUMN status TYPE video_status USING status::video_status;
ALTER TABLE videos ALTER COLUMN status SET DEFAULT 'draft';
-- +goose StatementEnd
