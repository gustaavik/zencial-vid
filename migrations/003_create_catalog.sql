-- +goose Up
CREATE TABLE genres (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE content_genres (
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, genre_id)
);

CREATE TABLE content_tags (
    content_id UUID NOT NULL REFERENCES content(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, tag_id)
);

-- Seed some default genres
INSERT INTO genres (id, name, slug) VALUES
    (uuid_generate_v4(), 'Action', 'action'),
    (uuid_generate_v4(), 'Comedy', 'comedy'),
    (uuid_generate_v4(), 'Drama', 'drama'),
    (uuid_generate_v4(), 'Horror', 'horror'),
    (uuid_generate_v4(), 'Sci-Fi', 'sci-fi'),
    (uuid_generate_v4(), 'Thriller', 'thriller'),
    (uuid_generate_v4(), 'Romance', 'romance'),
    (uuid_generate_v4(), 'Documentary', 'documentary'),
    (uuid_generate_v4(), 'Animation', 'animation'),
    (uuid_generate_v4(), 'Fantasy', 'fantasy');

-- +goose Down
DROP TABLE IF EXISTS content_tags;
DROP TABLE IF EXISTS content_genres;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS genres;
