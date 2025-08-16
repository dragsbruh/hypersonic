package storage

import (
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/dragsbruh/hypersonic/internal/config"
)

var ImageResolutions = []int{512}

func StoreCoverArt(cover io.Reader, albumHash string) error {
	var img image.Image
	for _, res := range ImageResolutions {
		targetPath := filepath.Join(config.DATA_DIR, "albums", albumHash, strconv.Itoa(res)+".webp")
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			if img == nil {
				img, _, err = image.Decode(cover)
				if err != nil {
					return fmt.Errorf("decode image: %w", err)
				}
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}

			file, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			defer file.Close()

			resized := imaging.Resize(img, res, res, imaging.Lanczos)
			if err := webp.Encode(file, resized, nil); err != nil {
				return fmt.Errorf("encode webp: %w", err)
			}
		}
	}

	return nil
}
