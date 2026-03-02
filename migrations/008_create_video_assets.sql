-- +migrate Up
CREATE TABLE IF NOT EXISTS video_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    qualities JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One asset per content (when not tied to an episode)
CREATE UNIQUE INDEX IF NOT EXISTS idx_video_assets_content_no_episode
    ON video_assets (content_id) WHERE episode_id IS NULL;

-- One asset per episode
CREATE UNIQUE INDEX IF NOT EXISTS idx_video_assets_episode
    ON video_assets (episode_id) WHERE episode_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_video_assets_content_id ON video_assets (content_id);
CREATE INDEX IF NOT EXISTS idx_video_assets_status ON video_assets (status);

-- +migrate Down
DROP TABLE IF EXISTS video_assets;
