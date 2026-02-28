package valueobject

// VideoQuality represents a video quality level.
type VideoQuality string

const (
	QualitySD  VideoQuality = "SD"  // 480p
	QualityHD  VideoQuality = "HD"  // 720p
	QualityFHD VideoQuality = "FHD" // 1080p
	QualityUHD VideoQuality = "UHD" // 4K
)

// Rank returns a numeric rank for comparison (higher is better).
func (q VideoQuality) Rank() int {
	switch q {
	case QualitySD:
		return 1
	case QualityHD:
		return 2
	case QualityFHD:
		return 3
	case QualityUHD:
		return 4
	default:
		return 0
	}
}

// IsAtLeast checks if this quality is at least as good as the given quality.
func (q VideoQuality) IsAtLeast(other VideoQuality) bool {
	return q.Rank() >= other.Rank()
}

// IsValid checks if the quality is a known value.
func (q VideoQuality) IsValid() bool {
	return q.Rank() > 0
}

// String returns the string representation.
func (q VideoQuality) String() string {
	return string(q)
}
