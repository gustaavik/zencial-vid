package pagination

import "github.com/zenfulcode/zencial/internal/pkg/httputil"

// FromParams creates pagination Meta from total count and page info.
func NewMeta(page, perPage int, total int64) *httputil.Meta {
	totalPages := 0
	if total > 0 {
		totalPages = int(total) / perPage
		if int(total)%perPage > 0 {
			totalPages++
		}
	}
	return &httputil.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
