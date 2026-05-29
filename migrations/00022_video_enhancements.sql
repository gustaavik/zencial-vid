-- +goose Up
CREATE TYPE submission_status AS ENUM ('draft', 'submitted', 'under_review', 'approved', 'rejected');
CREATE TYPE geo_restriction_type AS ENUM ('worldwide', 'include', 'exclude');

ALTER TABLE videos
    ADD COLUMN logline               VARCHAR(160)         NOT NULL DEFAULT '',
    ADD COLUMN primary_language      VARCHAR(10)          NOT NULL DEFAULT 'en',
    ADD COLUMN visibility            video_visibility     NOT NULL DEFAULT 'public',
    ADD COLUMN scheduled_publish_at  TIMESTAMPTZ,
    ADD COLUMN monetization_types    JSONB                NOT NULL DEFAULT '[]',
    ADD COLUMN ppv_price_cents       INT,
    ADD COLUMN free_preview_seconds  INT,
    ADD COLUMN ad_break_positions    JSONB                NOT NULL DEFAULT '[]',
    ADD COLUMN geo_restriction_type  geo_restriction_type NOT NULL DEFAULT 'worldwide',
    ADD COLUMN geo_restriction_regions JSONB              NOT NULL DEFAULT '[]',
    ADD COLUMN require_signin        BOOLEAN              NOT NULL DEFAULT FALSE,
    ADD COLUMN submission_status     submission_status    NOT NULL DEFAULT 'draft',
    ADD COLUMN submitted_at          TIMESTAMPTZ,
    ADD COLUMN moderator_notes       TEXT                 NOT NULL DEFAULT '',
    ADD COLUMN thumbnail_candidates  JSONB                NOT NULL DEFAULT '[]';

CREATE INDEX idx_videos_submission_status ON videos(submission_status);
CREATE INDEX idx_videos_visibility ON videos(visibility);
CREATE INDEX idx_videos_scheduled_publish_at ON videos(scheduled_publish_at) WHERE scheduled_publish_at IS NOT NULL;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_trusted_publisher BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
DROP INDEX IF EXISTS idx_videos_scheduled_publish_at;
DROP INDEX IF EXISTS idx_videos_visibility;
DROP INDEX IF EXISTS idx_videos_submission_status;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_trusted_publisher;

ALTER TABLE videos
    DROP COLUMN IF EXISTS logline,
    DROP COLUMN IF EXISTS primary_language,
    DROP COLUMN IF EXISTS visibility,
    DROP COLUMN IF EXISTS scheduled_publish_at,
    DROP COLUMN IF EXISTS monetization_types,
    DROP COLUMN IF EXISTS ppv_price_cents,
    DROP COLUMN IF EXISTS free_preview_seconds,
    DROP COLUMN IF EXISTS ad_break_positions,
    DROP COLUMN IF EXISTS geo_restriction_type,
    DROP COLUMN IF EXISTS geo_restriction_regions,
    DROP COLUMN IF EXISTS require_signin,
    DROP COLUMN IF EXISTS submission_status,
    DROP COLUMN IF EXISTS submitted_at,
    DROP COLUMN IF EXISTS moderator_notes,
    DROP COLUMN IF EXISTS thumbnail_candidates;

DROP TYPE IF EXISTS geo_restriction_type;
DROP TYPE IF EXISTS submission_status;
