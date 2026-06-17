-- +goose Up
ALTER TABLE series
    ADD COLUMN content_rating VARCHAR(20) NOT NULL DEFAULT '';

-- Widen origin_country to hold free-text country names (e.g. "United States").
ALTER TABLE series
    ALTER COLUMN origin_country TYPE VARCHAR(100);

-- +goose Down
ALTER TABLE series
    ALTER COLUMN origin_country TYPE VARCHAR(10);

ALTER TABLE series
    DROP COLUMN IF EXISTS content_rating;
