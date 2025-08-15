package media

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"

	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/dragsbruh/hypersonic/internal/config"
)

var ImageResolutions = []int{128, 256, 512}

func StoreAlbumArt(hash string, raw []byte) error {
	pathFunc := func(res int) string {
		return fmt.Sprintf("%s/albums/%d/%s.webp", config.DataDir, res, hash)
	}

	if !isAnyResMissing(pathFunc) {
		return nil
	}

	srcImage, _, err := image.Decode(bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}

	if err := generateRes(srcImage, pathFunc); err != nil {
		return fmt.Errorf("generate res: %w", err)
	}

	return nil
}

func isAnyResMissing(pathFunc func(int) string) bool {
	return slices.ContainsFunc(ImageResolutions, func(res int) bool {
		_, err := os.Stat(pathFunc(res))
		return os.IsNotExist(err)
	})
}

func generateRes(srcImage image.Image, pathFunc func(int) string) error {
	for _, res := range ImageResolutions {
		if err := os.MkdirAll(filepath.Dir(pathFunc(res)), os.ModePerm); err != nil {
			return fmt.Errorf("mkdir dir for res: %w", err)
		}

		resized := imaging.Resize(srcImage, res, res, imaging.Lanczos)
		if err := webp.Save(pathFunc(res), resized, nil); err != nil {
			return fmt.Errorf("save webp: %w", err)
		}
	}

	return nil
}
