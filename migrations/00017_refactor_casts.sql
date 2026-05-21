-- +goose Up

-- 1. Create the standalone casts table.
CREATE TABLE casts (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    picture_key VARCHAR(500),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (name)
);

-- 2. Populate casts from the distinct names already in video_cast.
--    For each name, carry over the most recent picture_key (if any).
INSERT INTO casts (id, name, picture_key, created_at, updated_at)
SELECT
    gen_random_uuid(),
    name,
    (
        SELECT picture_key
        FROM video_cast vc2
        WHERE vc2.name = vc.name AND vc2.picture_key IS NOT NULL
        ORDER BY vc2.updated_at DESC
        LIMIT 1
    ),
    MIN(created_at),
    MAX(updated_at)
FROM video_cast vc
GROUP BY name;

-- 3. Add cast_id to video_cast and populate it from the name match.
ALTER TABLE video_cast ADD COLUMN cast_id UUID;

UPDATE video_cast vc
SET cast_id = c.id
FROM casts c
WHERE c.name = vc.name;

ALTER TABLE video_cast
    ALTER COLUMN cast_id SET NOT NULL,
    ADD CONSTRAINT fk_video_cast_cast FOREIGN KEY (cast_id) REFERENCES casts (id) ON DELETE CASCADE;

CREATE INDEX idx_video_cast_cast_id ON video_cast (cast_id);

-- 4. Add unique constraint to prevent duplicate credits per video.
ALTER TABLE video_cast ADD CONSTRAINT uq_video_cast_video_cast UNIQUE (video_id, cast_id);

-- 5. Drop the columns that are now on the casts table.
ALTER TABLE video_cast DROP COLUMN name;
ALTER TABLE video_cast DROP COLUMN picture_key;

-- +goose Down

-- Restore name and picture_key columns on video_cast.
ALTER TABLE video_cast ADD COLUMN name        VARCHAR(255);
ALTER TABLE video_cast ADD COLUMN picture_key VARCHAR(500);

-- Populate them from the casts table before dropping it.
UPDATE video_cast vc
SET
    name        = c.name,
    picture_key = c.picture_key
FROM casts c
WHERE c.id = vc.cast_id;

ALTER TABLE video_cast ALTER COLUMN name SET NOT NULL;

-- Drop constraints and columns added in Up.
ALTER TABLE video_cast DROP CONSTRAINT IF EXISTS uq_video_cast_video_cast;
DROP INDEX IF EXISTS idx_video_cast_cast_id;
ALTER TABLE video_cast DROP CONSTRAINT IF EXISTS fk_video_cast_cast;
ALTER TABLE video_cast DROP COLUMN cast_id;

DROP TABLE IF EXISTS casts;
