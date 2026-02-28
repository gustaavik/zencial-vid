package valueobject

// ContentRating represents a content age rating.
type ContentRating string

const (
	RatingG    ContentRating = "G"    // General audiences
	RatingPG   ContentRating = "PG"   // Parental guidance suggested
	RatingPG13 ContentRating = "PG13" // Parents strongly cautioned
	RatingR    ContentRating = "R"    // Restricted (17+)
	RatingNC17 ContentRating = "NC17" // Adults only (18+)
)

// MinAge returns the minimum age required to view this content.
func (r ContentRating) MinAge() int {
	switch r {
	case RatingG:
		return 0
	case RatingPG:
		return 0
	case RatingPG13:
		return 13
	case RatingR:
		return 17
	case RatingNC17:
		return 18
	default:
		return 0
	}
}

// IsUnrestricted reports whether the rating has no age restriction.
func (r ContentRating) IsUnrestricted() bool {
	return r.MinAge() == 0
}

// AllowedForAge checks if a person of the given age can view this content.
func (r ContentRating) AllowedForAge(age int) bool {
	return age >= r.MinAge()
}

// IsValid checks if the rating is a known value.
func (r ContentRating) IsValid() bool {
	switch r {
	case RatingG, RatingPG, RatingPG13, RatingR, RatingNC17:
		return true
	}
	return false
}

// String returns the string representation.
func (r ContentRating) String() string {
	return string(r)
}
