-- +goose Up
ALTER TABLE video_cast DROP CONSTRAINT uq_video_cast_video_cast;
ALTER TABLE video_cast ADD CONSTRAINT uq_video_cast_video_cast_role UNIQUE (video_id, cast_id, role);

-- +goose Down
ALTER TABLE video_cast DROP CONSTRAINT uq_video_cast_video_cast_role;
ALTER TABLE video_cast ADD CONSTRAINT uq_video_cast_video_cast UNIQUE (video_id, cast_id);
