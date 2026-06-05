-- +goose Up
CREATE TYPE music_use_type AS ENUM ('original_score', 'needle_drop', 'sync_license', 'background');
CREATE TYPE music_rights_status AS ENUM ('owned', 'pending_clearance', 'cleared', 'rejected');

CREATE TABLE music_cues (
    id                    UUID                PRIMARY KEY,
    video_id              UUID                NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    timecode_seconds      INT                 NOT NULL,
    title                 VARCHAR(500)        NOT NULL,
    composer_artist       VARCHAR(500)        NOT NULL DEFAULT '',
    use_type              music_use_type      NOT NULL DEFAULT 'original_score',
    rights_status         music_rights_status NOT NULL DEFAULT 'owned',
    clearance_document_key TEXT               NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ         NOT NULL,
    updated_at            TIMESTAMPTZ         NOT NULL
);

CREATE INDEX idx_music_cues_video_id ON music_cues(video_id, timecode_seconds);

-- +goose Down
DROP TABLE IF EXISTS music_cues;
DROP TYPE IF EXISTS music_rights_status;
DROP TYPE IF EXISTS music_use_type;
