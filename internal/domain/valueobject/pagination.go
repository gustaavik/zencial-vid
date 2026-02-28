package valueobject

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Pagination represents pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
}

// NewPagination creates validated pagination parameters.
func NewPagination(page, perPage int) Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return Pagination{Page: page, PerPage: perPage}
}

// Offset returns the SQL OFFSET value.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL LIMIT value.
func (p Pagination) Limit() int {
	return p.PerPage
}

// TotalPages calculates the total number of pages.
func (p Pagination) TotalPages(totalItems int64) int {
	if totalItems == 0 {
		return 0
	}
	pages := int(totalItems) / p.PerPage
	if int(totalItems)%p.PerPage > 0 {
		pages++
	}
	return pages
}
