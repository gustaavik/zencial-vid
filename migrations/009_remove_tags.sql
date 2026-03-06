-- +goose Up
DROP TABLE IF EXISTS content_tags;
DROP TABLE IF EXISTS tags;

-- +goose Down
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE content_tags (
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    tag_id     UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, tag_id)
);
