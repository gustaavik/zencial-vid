-- +goose Up
-- Migration 010: typed entities
-- Adds genre_id and plan_id FKs to content rows (Film and Video),
-- uploaded_at to videos, trailer_url/backdrop_url to seasons,
-- series_id/director to episodes, and drops the content_genres join table.

-- Add single genre FK to content (replaces many-to-many content_genres)
ALTER TABLE content ADD COLUMN IF NOT EXISTS genre_id UUID REFERENCES genres(id) ON DELETE SET NULL;

-- Add plan FK to content (NULL = free to all subscribers)
ALTER TABLE content ADD COLUMN IF NOT EXISTS plan_id UUID REFERENCES plans(id) ON DELETE SET NULL;

-- Video: record when the video was uploaded
ALTER TABLE videos ADD COLUMN IF NOT EXISTS uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Season: media URLs for season-level trailers and backdrops
ALTER TABLE seasons ADD COLUMN IF NOT EXISTS trailer_url  TEXT NOT NULL DEFAULT '';
ALTER TABLE seasons ADD COLUMN IF NOT EXISTS backdrop_url TEXT NOT NULL DEFAULT '';

-- Episode: denormalised series reference for direct episode queries
ALTER TABLE episodes ADD COLUMN IF NOT EXISTS series_id UUID REFERENCES content(id) ON DELETE CASCADE;

-- Episode: director credit
ALTER TABLE episodes ADD COLUMN IF NOT EXISTS director VARCHAR(255) NOT NULL DEFAULT '';

-- Populate series_id on existing episodes via season → content join
UPDATE episodes e
SET series_id = s.content_id
FROM seasons s
WHERE e.season_id = s.id
  AND e.series_id IS NULL;

-- Drop the many-to-many genre join table (replaced by content.genre_id).
-- NOTE: existing genre assignments will be lost; reassign via API.
DROP TABLE IF EXISTS content_genres;

-- +goose Down
ALTER TABLE content DROP COLUMN IF EXISTS genre_id;
ALTER TABLE content DROP COLUMN IF EXISTS plan_id;
ALTER TABLE videos DROP COLUMN IF EXISTS uploaded_at;
ALTER TABLE seasons DROP COLUMN IF EXISTS trailer_url;
ALTER TABLE seasons DROP COLUMN IF EXISTS backdrop_url;
ALTER TABLE episodes DROP COLUMN IF EXISTS series_id;
ALTER TABLE episodes DROP COLUMN IF EXISTS director;
