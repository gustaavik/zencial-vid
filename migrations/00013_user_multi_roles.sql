-- +goose Up
ALTER TABLE users ADD COLUMN roles TEXT[] NOT NULL DEFAULT ARRAY['user'];
UPDATE users SET roles = ARRAY[role::text];
ALTER TABLE users DROP COLUMN role;

-- +goose Down
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
UPDATE users SET role = roles[1];
ALTER TABLE users DROP COLUMN roles;
