-- +goose Up
CREATE TABLE video_cast (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id   UUID        NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    role       VARCHAR(100) NOT NULL,
    sort_order INT         NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_video_cast_video_id ON video_cast(video_id);

-- +goose Down
DROP TABLE IF EXISTS video_cast;
