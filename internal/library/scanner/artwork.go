package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dragsbruh/hypersonic/internal/config"
)

var ArtworkResolutions = []int{128, 256, 512}

func ExtractArtwork(albumHash, trackPath string) error {
	dir := filepath.Join(config.DataDir, "albums", albumHash)
	tempDir := filepath.Join(config.DataDir, "tmp")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	for _, res := range ArtworkResolutions {
		imagePath := filepath.Join(dir, fmt.Sprintf("%d.webp", res))
		_, err := os.Stat(imagePath)
		if err == nil {
			continue
		}

		tmpPath := filepath.Join(tempDir, fmt.Sprintf("%s_%d.webp", albumHash, res))

		cmd := exec.Command(
			"ffmpeg",
			"-hide_banner",
			"-loglevel", "error",
			"-i", trackPath,
			"-vf", fmt.Sprintf("scale=%d:%d", res, res),
			"-c:v", "libwebp",
			tmpPath,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("ffmpeg error: %w, output: %s", err, string(out))
		}
		if err := os.Rename(tmpPath, imagePath); err != nil {
			return fmt.Errorf("rename: %v", err)
		}
	}
	return nil
}
