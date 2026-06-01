-- +goose Up
CREATE TYPE chapter_source AS ENUM ('auto', 'manual');

CREATE TABLE chapters (
    id               UUID           PRIMARY KEY,
    video_id         UUID           NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    start_time_secs  INT            NOT NULL,
    title            VARCHAR(500)   NOT NULL,
    source           chapter_source NOT NULL DEFAULT 'manual',
    created_at       TIMESTAMPTZ    NOT NULL,
    updated_at       TIMESTAMPTZ    NOT NULL
);

CREATE INDEX idx_chapters_video_id ON chapters(video_id, start_time_secs);

-- +goose Down
DROP TABLE IF EXISTS chapters;
DROP TYPE IF EXISTS chapter_source;
