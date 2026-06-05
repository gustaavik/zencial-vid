-- +goose Up
ALTER TABLE videos
    ADD COLUMN is_featured          BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN featured_description TEXT,
    ADD COLUMN featured_at          TIMESTAMPTZ;

CREATE INDEX idx_videos_is_featured ON videos (featured_at DESC) WHERE is_featured = TRUE;

-- +goose Down
DROP INDEX IF EXISTS idx_videos_is_featured;
ALTER TABLE videos
    DROP COLUMN IF EXISTS featured_at,
    DROP COLUMN IF EXISTS featured_description,
    DROP COLUMN IF EXISTS is_featured;
