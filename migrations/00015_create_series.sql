-- +goose Up
CREATE TYPE series_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE series (
    id                UUID         PRIMARY KEY,
    title             VARCHAR(500) NOT NULL,
    slug              VARCHAR(500) NOT NULL UNIQUE,
    description       TEXT         NOT NULL DEFAULT '',
    creator           VARCHAR(200) NOT NULL DEFAULT '',
    status            series_status NOT NULL DEFAULT 'draft',
    cover_image_key   TEXT         NOT NULL DEFAULT '',
    uploaded_by       UUID         NOT NULL REFERENCES users(id),
    minimum_plan_level INT,
    created_at        TIMESTAMPTZ  NOT NULL,
    updated_at        TIMESTAMPTZ  NOT NULL
);

CREATE INDEX idx_series_status ON series(status);
CREATE INDEX idx_series_uploaded_by ON series(uploaded_by);

CREATE TABLE series_genres (
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    genre_id  UUID NOT NULL REFERENCES genres(id)  ON DELETE CASCADE,
    PRIMARY KEY (series_id, genre_id)
);

ALTER TABLE videos
    ADD COLUMN series_id      UUID REFERENCES series(id) ON DELETE SET NULL,
    ADD COLUMN season_number  INT,
    ADD COLUMN episode_number INT;

CREATE INDEX idx_videos_series_id ON videos(series_id);
CREATE UNIQUE INDEX idx_videos_series_episode
    ON videos(series_id, season_number, episode_number)
    WHERE series_id IS NOT NULL;

CREATE TABLE series_watch_progress (
    user_id         UUID        NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    series_id       UUID        NOT NULL REFERENCES series(id)  ON DELETE CASCADE,
    last_episode_id UUID        NOT NULL REFERENCES videos(id)  ON DELETE CASCADE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, series_id)
);

CREATE INDEX idx_series_watch_progress_user ON series_watch_progress(user_id, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_videos_series_episode;
DROP INDEX IF EXISTS idx_videos_series_id;
ALTER TABLE videos
    DROP COLUMN IF EXISTS series_id,
    DROP COLUMN IF EXISTS season_number,
    DROP COLUMN IF EXISTS episode_number;
DROP TABLE IF EXISTS series_watch_progress;
DROP TABLE IF EXISTS series_genres;
DROP TABLE IF EXISTS series;
DROP TYPE IF EXISTS series_status;
