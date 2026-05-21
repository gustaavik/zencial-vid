-- +goose Up
ALTER TABLE video_cast ADD COLUMN picture_key VARCHAR(500);

-- +goose Down
ALTER TABLE video_cast DROP COLUMN picture_key;
