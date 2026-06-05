-- +goose Up
CREATE TYPE caption_status AS ENUM ('pending', 'auto_generated', 'reviewed', 'published');
CREATE TYPE caption_source AS ENUM ('auto', 'manual');

CREATE TABLE captions (
    id            UUID           PRIMARY KEY,
    video_id      UUID           NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    language_code VARCHAR(10)    NOT NULL,
    format        VARCHAR(20)    NOT NULL DEFAULT 'webvtt',
    storage_key   TEXT           NOT NULL DEFAULT '',
    status        caption_status NOT NULL DEFAULT 'pending',
    source        caption_source NOT NULL DEFAULT 'manual',
    created_at    TIMESTAMPTZ    NOT NULL,
    updated_at    TIMESTAMPTZ    NOT NULL,
    UNIQUE (video_id, language_code)
);

CREATE INDEX idx_captions_video_id ON captions(video_id);

-- +goose Down
DROP TABLE IF EXISTS captions;
DROP TYPE IF EXISTS caption_source;
DROP TYPE IF EXISTS caption_status;
