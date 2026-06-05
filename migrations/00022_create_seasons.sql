-- +goose Up
CREATE TYPE release_cadence AS ENUM ('all_at_once', 'weekly', 'bi_weekly', 'on_demand');

CREATE TABLE seasons (
    id               UUID           PRIMARY KEY,
    series_id        UUID           NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    season_number    INT            NOT NULL,
    season_tag       VARCHAR(200)   NOT NULL DEFAULT '',
    planned_episodes INT            NOT NULL DEFAULT 0,
    avg_runtime_secs INT            NOT NULL DEFAULT 0,
    release_cadence  release_cadence NOT NULL DEFAULT 'on_demand',
    premiere_date    TIMESTAMPTZ,
    cadence_day      INT,
    timezone         VARCHAR(100)   NOT NULL DEFAULT 'UTC',
    created_at       TIMESTAMPTZ    NOT NULL,
    updated_at       TIMESTAMPTZ    NOT NULL,
    UNIQUE (series_id, season_number),
    CONSTRAINT cadence_day_range CHECK (cadence_day IS NULL OR (cadence_day >= 0 AND cadence_day <= 6))
);

CREATE INDEX idx_seasons_series_id ON seasons(series_id);

-- +goose Down
DROP TABLE IF EXISTS seasons;
DROP TYPE IF EXISTS release_cadence;
