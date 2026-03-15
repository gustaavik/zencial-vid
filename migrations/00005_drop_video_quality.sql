-- +goose Up
ALTER TABLE videos DROP COLUMN quality;

-- +goose Down
ALTER TABLE videos ADD COLUMN quality VARCHAR(10) NOT NULL DEFAULT 'HD';
