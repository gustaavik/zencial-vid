-- +goose Up
CREATE TABLE content (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    type VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    synopsis TEXT NOT NULL DEFAULT '',
    rating VARCHAR(10) NOT NULL DEFAULT 'G',
    release_year INT NOT NULL DEFAULT 0,
    poster_url TEXT NOT NULL DEFAULT '',
    backdrop_url TEXT NOT NULL DEFAULT '',
    trailer_url TEXT NOT NULL DEFAULT '',
    director VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_content_slug ON content (slug);

CREATE INDEX idx_content_status ON content (status);

CREATE INDEX idx_content_type ON content(type);

CREATE INDEX idx_content_featured ON content (is_featured)
WHERE
    is_featured = TRUE;

CREATE INDEX idx_content_search ON content USING GIN (
    to_tsvector(
        'english',
        title || ' ' || description
    )
);

CREATE TABLE films (
    content_id UUID PRIMARY KEY REFERENCES content (id) ON DELETE CASCADE,
    duration_seconds BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE seasons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    content_id UUID NOT NULL REFERENCES content (id) ON DELETE CASCADE,
    number INT NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (content_id, number)
);

CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    season_id UUID NOT NULL REFERENCES seasons (id) ON DELETE CASCADE,
    number INT NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    synopsis TEXT NOT NULL DEFAULT '',
    duration_seconds BIGINT NOT NULL DEFAULT 0,
    air_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (season_id, number)
);

CREATE TABLE video_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    content_id UUID REFERENCES content (id) ON DELETE SET NULL,
    episode_id UUID REFERENCES episodes (id) ON DELETE SET NULL,
    storage_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS video_assets;

DROP TABLE IF EXISTS episodes;

DROP TABLE IF EXISTS seasons;

DROP TABLE IF EXISTS films;

DROP TABLE IF EXISTS content;