package mapper

import (
	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
)

// ContentToListResponse maps a Content entity to a list DTO.
func ContentToListResponse(c *entity.Content) dto.ContentListResponse {
	genres := make([]string, len(c.Genres))
	for i, g := range c.Genres {
		genres[i] = g.Name
	}
	resp := dto.ContentListResponse{
		ID:          c.ID.String(),
		Type:        string(c.Type),
		Title:       c.Title,
		Slug:        c.Slug.String(),
		Description: c.Description,
		Rating:      string(c.Rating),
		ReleaseYear: c.ReleaseYear,
		PosterURL:   c.PosterURL,
		Genres:      genres,
		IsFeatured:  c.IsFeatured,
	}
	if c.Video != nil {
		resp.CreatorName = c.Video.CreatorName
		resp.IsFree = c.Video.IsFree
	}
	return resp
}

// ContentToDetailResponse maps a Content entity to a detail DTO.
func ContentToDetailResponse(c *entity.Content) dto.ContentDetailResponse {
	genres := make([]dto.GenreResponse, len(c.Genres))
	for i, g := range c.Genres {
		genres[i] = GenreToResponse(&g)
	}

	cast := make([]dto.CastMemberResponse, len(c.Cast))
	for i, m := range c.Cast {
		cast[i] = dto.CastMemberResponse{
			Name:      m.Name,
			Role:      m.Role,
			Character: m.Character,
			ImageURL:  m.ImageURL,
		}
	}

	resp := dto.ContentDetailResponse{
		ID:          c.ID.String(),
		Type:        string(c.Type),
		Title:       c.Title,
		Slug:        c.Slug.String(),
		Description: c.Description,
		Synopsis:    c.Synopsis,
		Rating:      string(c.Rating),
		ReleaseYear: c.ReleaseYear,
		PosterURL:   c.PosterURL,
		BackdropURL: c.BackdropURL,
		TrailerURL:  c.TrailerURL,
		Director:    c.Director,
		IsFeatured:  c.IsFeatured,
		Genres:      genres,
		Cast:        cast,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if c.Film != nil {
		film := &dto.FilmResponse{
			DurationMinutes: c.Film.Duration.Minutes(),
		}
		if c.Film.Asset.ID != uuid.Nil {
			film.Asset = VideoAssetToResponse(&c.Film.Asset)
		}
		resp.Film = film
	}
	if c.Series != nil {
		resp.Series = &dto.SeriesResponse{
			TotalSeasons: c.Series.TotalSeasons,
		}
	}
	if c.Video != nil {
		video := &dto.VideoResponse{
			DurationMinutes: c.Video.Duration.Minutes(),
			CreatorName:     c.Video.CreatorName,
			IsFree:          c.Video.IsFree,
		}
		if c.Video.Asset.ID != uuid.Nil {
			video.Asset = VideoAssetToResponse(&c.Video.Asset)
		}
		resp.Video = video
	}

	return resp
}

// ContentsToListResponse maps a slice of Content entities to list DTOs.
func ContentsToListResponse(contents []entity.Content) []dto.ContentListResponse {
	result := make([]dto.ContentListResponse, len(contents))
	for i := range contents {
		result[i] = ContentToListResponse(&contents[i])
	}
	return result
}

// GenreToResponse maps a Genre entity to a DTO.
func GenreToResponse(g *entity.Genre) dto.GenreResponse {
	return dto.GenreResponse{
		ID:   g.ID.String(),
		Name: g.Name,
		Slug: g.Slug,
	}
}

// GenresToResponse maps a slice of Genre entities to DTOs.
func GenresToResponse(genres []entity.Genre) []dto.GenreResponse {
	result := make([]dto.GenreResponse, len(genres))
	for i := range genres {
		result[i] = GenreToResponse(&genres[i])
	}
	return result
}

// CategoryToResponse maps a Category entity to a DTO.
func CategoryToResponse(c *entity.Category) dto.CategoryResponse {
	resp := dto.CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
	}
	if c.ParentID != nil {
		pid := c.ParentID.String()
		resp.ParentID = &pid
	}
	return resp
}

// CategoriesToResponse maps a slice of Category entities to DTOs.
func CategoriesToResponse(categories []entity.Category) []dto.CategoryResponse {
	result := make([]dto.CategoryResponse, len(categories))
	for i := range categories {
		result[i] = CategoryToResponse(&categories[i])
	}
	return result
}

// SeasonToResponse maps a Season entity to a DTO.
func SeasonToResponse(s *entity.Season) dto.SeasonResponse {
	return dto.SeasonResponse{
		ID:       s.ID.String(),
		Number:   s.Number,
		Title:    s.Title,
		Episodes: len(s.Episodes),
	}
}

// SeasonsToResponse maps a slice of Season entities to DTOs.
func SeasonsToResponse(seasons []entity.Season) []dto.SeasonResponse {
	result := make([]dto.SeasonResponse, len(seasons))
	for i := range seasons {
		result[i] = SeasonToResponse(&seasons[i])
	}
	return result
}

// EpisodeToResponse maps an Episode entity to a DTO.
func EpisodeToResponse(e *entity.Episode) dto.EpisodeResponse {
	resp := dto.EpisodeResponse{
		ID:              e.ID.String(),
		Number:          e.Number,
		Title:           e.Title,
		Synopsis:        e.Synopsis,
		DurationMinutes: e.Duration.Minutes(),
	}
	if e.AirDate != nil {
		ad := e.AirDate.Format("2006-01-02")
		resp.AirDate = &ad
	}
	return resp
}

// EpisodesToResponse maps a slice of Episode entities to DTOs.
func EpisodesToResponse(episodes []entity.Episode) []dto.EpisodeResponse {
	result := make([]dto.EpisodeResponse, len(episodes))
	for i := range episodes {
		result[i] = EpisodeToResponse(&episodes[i])
	}
	return result
}

// VideoAssetToResponse maps a VideoAsset entity to a DTO.
func VideoAssetToResponse(a *entity.VideoAsset) *dto.VideoAssetResponse {
	qualities := make([]dto.VideoRenditionResponse, len(a.Qualities))
	for i, q := range a.Qualities {
		qualities[i] = dto.VideoRenditionResponse{
			Quality:    string(q.Quality),
			URL:        q.URL,
			Bitrate:    q.Bitrate,
			Resolution: q.Resolution,
		}
	}
	return &dto.VideoAssetResponse{
		ID:         a.ID.String(),
		StorageKey: a.StorageKey,
		Status:     string(a.Status),
		Qualities:  qualities,
	}
}
