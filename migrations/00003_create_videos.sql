-- +goose Up
CREATE TYPE video_status AS ENUM ('draft', 'processing', 'published', 'archived');

CREATE TABLE videos (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title          VARCHAR(500)   NOT NULL,
    slug           VARCHAR(500)   NOT NULL UNIQUE,
    description    TEXT           NOT NULL DEFAULT '',
    creator        VARCHAR(255)   NOT NULL DEFAULT '',
    duration       BIGINT         NOT NULL DEFAULT 0,
    content_rating VARCHAR(10)    NOT NULL DEFAULT 'G',
    quality        VARCHAR(10)    NOT NULL DEFAULT 'HD',
    status         video_status   NOT NULL DEFAULT 'draft',
    storage_key    VARCHAR(1024)  NOT NULL,
    content_type   VARCHAR(255)   NOT NULL DEFAULT 'video/mp4',
    file_size      BIGINT         NOT NULL DEFAULT 0,
    thumbnail_key  VARCHAR(1024)  NOT NULL DEFAULT '',
    uploaded_by    UUID           NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE video_genres (
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (video_id, genre_id)
);

CREATE INDEX idx_videos_slug ON videos(slug);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_uploaded_by ON videos(uploaded_by);
CREATE INDEX idx_video_genres_genre_id ON video_genres(genre_id);

-- +goose Down
DROP TABLE IF EXISTS video_genres;
DROP TABLE IF EXISTS videos;
DROP TYPE IF EXISTS video_status;
