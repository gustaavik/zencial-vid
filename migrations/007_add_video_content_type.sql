-- +goose Up
CREATE TABLE videos (
    content_id UUID PRIMARY KEY REFERENCES content (id) ON DELETE CASCADE,
    duration_seconds BIGINT NOT NULL DEFAULT 0,
    creator_name VARCHAR(255) NOT NULL DEFAULT '',
    is_free BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS videos;
