package thumbnail

import (
	"fmt"
	"os/exec"
)

// ExtractFirstFrame runs ffmpeg to extract the first frame of a video file
// and writes it as a JPEG to outputPath.
// The caller is responsible for cleaning up both inputPath and outputPath.
func ExtractFirstFrame(inputPath, outputPath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-vframes", "1",
		"-f", "image2",
		"-y",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract first frame: %w (output: %s)", err, output)
	}
	return nil
}
