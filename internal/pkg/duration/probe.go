package duration

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// Probe runs ffprobe on the given video file and returns its duration in seconds.
func Probe(videoPath string) (int64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration: %w", err)
	}

	raw := strings.TrimSpace(string(output))
	secs, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration: parse %q: %w", raw, err)
	}

	return int64(math.Round(secs)), nil
}
