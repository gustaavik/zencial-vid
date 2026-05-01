-- +goose Up
CREATE TABLE user_watch_progress (
    user_id          UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    video_id         UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    position_seconds BIGINT NOT NULL DEFAULT 0 CHECK (position_seconds >= 0),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, video_id)
);

CREATE INDEX idx_user_watch_progress_user_updated_at
    ON user_watch_progress(user_id, updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_watch_progress;
