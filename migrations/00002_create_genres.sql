-- +goose Up
CREATE TABLE genres (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE genre_translations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    genre_id      UUID         NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    language_code VARCHAR(5)   NOT NULL,
    name          VARCHAR(255) NOT NULL,
    description   TEXT         NOT NULL DEFAULT '',
    UNIQUE(genre_id, language_code)
);

CREATE INDEX idx_genre_translations_genre_id ON genre_translations(genre_id);
CREATE INDEX idx_genre_translations_lang ON genre_translations(language_code);

-- +goose Down
DROP TABLE IF EXISTS genre_translations;
DROP TABLE IF EXISTS genres;
