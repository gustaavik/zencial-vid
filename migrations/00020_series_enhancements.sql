-- +goose Up
CREATE TYPE series_type AS ENUM ('ongoing', 'limited', 'anthology', 'documentary');
CREATE TYPE video_visibility AS ENUM ('public', 'unlisted', 'followers_only', 'private');

ALTER TABLE series
    ADD COLUMN series_type          series_type  NOT NULL DEFAULT 'ongoing',
    ADD COLUMN logline              VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN primary_language     VARCHAR(10)  NOT NULL DEFAULT 'en',
    ADD COLUMN origin_country       VARCHAR(10)  NOT NULL DEFAULT '',
    ADD COLUMN poster_key           TEXT         NOT NULL DEFAULT '',
    ADD COLUMN banner_key           TEXT         NOT NULL DEFAULT '',
    ADD COLUMN title_logo_key       TEXT         NOT NULL DEFAULT '',
    ADD COLUMN autoplay_next        BOOLEAN      NOT NULL DEFAULT TRUE,
    ADD COLUMN binge_mode           BOOLEAN      NOT NULL DEFAULT TRUE,
    ADD COLUMN hide_episode_count   BOOLEAN      NOT NULL DEFAULT FALSE,
    ADD COLUMN default_visibility   video_visibility NOT NULL DEFAULT 'public',
    ADD COLUMN default_monetization JSONB        NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE series
    DROP COLUMN IF EXISTS series_type,
    DROP COLUMN IF EXISTS logline,
    DROP COLUMN IF EXISTS primary_language,
    DROP COLUMN IF EXISTS origin_country,
    DROP COLUMN IF EXISTS poster_key,
    DROP COLUMN IF EXISTS banner_key,
    DROP COLUMN IF EXISTS title_logo_key,
    DROP COLUMN IF EXISTS autoplay_next,
    DROP COLUMN IF EXISTS binge_mode,
    DROP COLUMN IF EXISTS hide_episode_count,
    DROP COLUMN IF EXISTS default_visibility,
    DROP COLUMN IF EXISTS default_monetization;

DROP TYPE IF EXISTS video_visibility;
DROP TYPE IF EXISTS series_type;
