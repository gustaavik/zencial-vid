-- +goose Up
CREATE TYPE cast_department AS ENUM (
    'performance', 'direction', 'cinematography', 'sound',
    'post', 'production', 'writing', 'vfx'
);
CREATE TYPE cast_invite_status AS ENUM ('not_invited', 'pending', 'accepted');

ALTER TABLE video_cast
    ADD COLUMN department    cast_department   NOT NULL DEFAULT 'performance',
    ADD COLUMN invite_status cast_invite_status NOT NULL DEFAULT 'not_invited',
    ADD COLUMN invited_email VARCHAR(500);

-- +goose Down
ALTER TABLE video_cast
    DROP COLUMN IF EXISTS department,
    DROP COLUMN IF EXISTS invite_status,
    DROP COLUMN IF EXISTS invited_email;

DROP TYPE IF EXISTS cast_invite_status;
DROP TYPE IF EXISTS cast_department;
