package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// GenreTranslation holds the localized name and description for a genre.
type GenreTranslation struct {
	ID           uuid.UUID
	GenreID      uuid.UUID
	LanguageCode valueobject.LanguageCode
	Name         string
	Description  string
}

// Genre is the core genre entity with multi-language support.
type Genre struct {
	ID           uuid.UUID
	Slug         valueobject.Slug
	Translations []GenreTranslation
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewGenre creates a new Genre from a slug.
func NewGenre(slug valueobject.Slug) *Genre {
	now := time.Now()
	return &Genre{
		ID:        uuid.New(),
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddTranslation adds or replaces a translation for the given language.
func (g *Genre) AddTranslation(langCode valueobject.LanguageCode, name, description string) {
	for i, t := range g.Translations {
		if t.LanguageCode.String() == langCode.String() {
			g.Translations[i].Name = name
			g.Translations[i].Description = description
			g.UpdatedAt = time.Now()
			return
		}
	}
	g.Translations = append(g.Translations, GenreTranslation{
		ID:           uuid.New(),
		GenreID:      g.ID,
		LanguageCode: langCode,
		Name:         name,
		Description:  description,
	})
	g.UpdatedAt = time.Now()
}

// TranslationFor returns the translation for a language, falling back to "en", then the first available.
func (g *Genre) TranslationFor(langCode string) *GenreTranslation {
	var fallback *GenreTranslation
	for i := range g.Translations {
		if g.Translations[i].LanguageCode.String() == langCode {
			return &g.Translations[i]
		}
		if g.Translations[i].LanguageCode.String() == "en" {
			fallback = &g.Translations[i]
		}
	}
	if fallback != nil {
		return fallback
	}
	if len(g.Translations) > 0 {
		return &g.Translations[0]
	}
	return nil
}
