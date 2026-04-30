-- +goose Up
CREATE TABLE user_watchlist (
    user_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, video_id)
);

CREATE INDEX idx_user_watchlist_user_added_at ON user_watchlist(user_id, added_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_watchlist;
