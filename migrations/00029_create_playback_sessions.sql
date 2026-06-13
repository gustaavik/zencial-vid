-- +goose Up
CREATE TABLE playback_sessions (
    id              UUID        PRIMARY KEY,
    video_id        UUID        NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id         UUID        REFERENCES users(id) ON DELETE SET NULL,
    source          VARCHAR(20) NOT NULL DEFAULT 'other',
    platform        VARCHAR(20) NOT NULL DEFAULT 'other',
    country_code    VARCHAR(2)  NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_event_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_position   BIGINT      NOT NULL DEFAULT 0 CHECK (last_position >= 0),
    max_position    BIGINT      NOT NULL DEFAULT 0 CHECK (max_position >= 0),
    watched_seconds BIGINT      NOT NULL DEFAULT 0 CHECK (watched_seconds >= 0),
    watched_buckets BIT(100)    NOT NULL DEFAULT (repeat('0', 100))::bit(100),
    is_view         BOOLEAN     NOT NULL DEFAULT FALSE,
    completed       BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_playback_sessions_video_started ON playback_sessions(video_id, started_at DESC);
CREATE INDEX idx_playback_sessions_video_views ON playback_sessions(video_id, started_at DESC) WHERE is_view;
CREATE INDEX idx_playback_sessions_user ON playback_sessions(user_id, started_at DESC) WHERE user_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS playback_sessions;
