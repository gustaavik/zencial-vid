package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	"github.com/zenfulcode/zencial/internal/domain/entity"
	contentuc "github.com/zenfulcode/zencial/internal/usecase/content"
)

// ContentSummaryToListResponse maps a ContentSummary to a list DTO.
func ContentSummaryToListResponse(c *entity.ContentSummary) dto.ContentListResponse {
	resp := dto.ContentListResponse{
		ID:          c.ID.String(),
		Type:        string(c.Type),
		Title:       c.Title,
		Slug:        c.Slug.String(),
		Description: c.Description,
		Status:      string(c.Status),
		Rating:      string(c.Rating),
		PosterURL:   c.PosterURL,
		IsFeatured:  c.IsFeatured,
		IsFree:      c.IsFree(),
		CreatorName: c.CreatorName,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.Genre != nil {
		g := GenreToResponse(c.Genre)
		resp.Genre = &g
	}
	return resp
}

// ContentSummariesToListResponse maps a slice of ContentSummary to list DTOs.
func ContentSummariesToListResponse(summaries []entity.ContentSummary) []dto.ContentListResponse {
	result := make([]dto.ContentListResponse, len(summaries))
	for i := range summaries {
		result[i] = ContentSummaryToListResponse(&summaries[i])
	}
	return result
}

// ContentDetailToResponse maps a ContentDetail discriminated union to a detail DTO.
func ContentDetailToResponse(d *contentuc.ContentDetail) dto.ContentDetailResponse {
	switch d.Type {
	case entity.ContentTypeFilm:
		return FilmToDetailResponse(d.Film)
	case entity.ContentTypeVideo:
		return VideoToDetailResponse(d.Video)
	case entity.ContentTypeSeries:
		return SeriesToDetailResponse(d.Series)
	default:
		return dto.ContentDetailResponse{}
	}
}

// FilmToDetailResponse maps a Film entity to a detail DTO.
func FilmToDetailResponse(f *entity.Film) dto.ContentDetailResponse {
	cast := make([]dto.CastMemberResponse, len(f.CastMembers))
	for i, m := range f.CastMembers {
		cast[i] = dto.CastMemberResponse{
			Name:      m.Name,
			Role:      m.Role,
			Character: m.Character,
			ImageURL:  m.ImageURL,
		}
	}
	resp := dto.ContentDetailResponse{
		ID:          f.ID.String(),
		Type:        string(f.Type),
		Title:       f.Title,
		Slug:        f.Slug.String(),
		Description: f.Description,
		Synopsis:    f.Synopsis,
		Rating:      string(f.Rating),
		PosterURL:   f.PosterURL,
		BackdropURL: f.BackdropURL,
		TrailerURL:  f.TrailerURL,
		Director:    f.Director,
		ReleaseYear: f.ReleaseYear,
		IsFeatured:  f.IsFeatured,
		IsFree:      f.IsFree(),
		Cast:        cast,
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if f.Genre != nil {
		g := GenreToResponse(f.Genre)
		resp.Genre = &g
	}
	if f.Asset != nil {
		resp.Asset = VideoAssetToResponse(f.Asset)
	}
	return resp
}

// VideoToDetailResponse maps a Video entity to a detail DTO.
func VideoToDetailResponse(v *entity.Video) dto.ContentDetailResponse {
	resp := dto.ContentDetailResponse{
		ID:          v.ID.String(),
		Type:        string(v.Type),
		Title:       v.Title,
		Slug:        v.Slug.String(),
		Description: v.Description,
		Synopsis:    v.Synopsis,
		Rating:      string(v.Rating),
		PosterURL:   v.PosterURL,
		IsFeatured:  v.IsFeatured,
		IsFree:      v.IsFree(),
		CreatorName: v.CreatorName,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if v.Genre != nil {
		g := GenreToResponse(v.Genre)
		resp.Genre = &g
	}
	if v.Asset != nil {
		resp.Asset = VideoAssetToResponse(v.Asset)
	}
	return resp
}

// SeriesToDetailResponse maps a Series entity to a detail DTO.
func SeriesToDetailResponse(s *entity.Series) dto.ContentDetailResponse {
	return dto.ContentDetailResponse{
		ID:           s.ID.String(),
		Type:         string(entity.ContentTypeSeries),
		Title:        s.Title,
		Slug:         s.Slug.String(),
		Description:  s.Description,
		Synopsis:     s.Synopsis,
		PosterURL:    s.PosterURL,
		BackdropURL:  s.BackdropURL,
		TrailerURL:   s.TrailerURL,
		IsFeatured:   s.IsFeatured,
		TotalSeasons: s.TotalSeasons,
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
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
