package library

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/mozillazg/go-unidecode"
)

type HypersonicArtist struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
}

type HypersonicAlbum struct {
	Hash string `json:"hash"`
	Name string `json:"name"`

	Artists []HypersonicArtist `json:"artists"`

	TotalTracks int `json:"totalTracks"`
	TotalDiscs  int `json:"totalDiscs"`
}

type HypersonicTrack struct {
	Hash string `json:"hash"`
	Name string `json:"name"`

	Artists []HypersonicArtist `json:"artists"`

	Album       *HypersonicAlbum `json:"album"`
	DiscNumber  int              `json:"discNumber"`
	TrackNumber int              `json:"trackNumber"`
	Year        int              `json:"year"`
	Genre       string           `json:"genre"`

	FilePath string `json:"filePath"`
	FileHash string `json:"fileHash"`
}

func CalculateHash(args ...string) string {
	var b strings.Builder
	for _, arg := range args {
		for _, char := range unidecode.Unidecode(arg) {
			if unicode.IsLetter(char) || unicode.IsDigit(char) {
				b.WriteRune(unicode.ToLower(char))
			}
		}
	}
	normalized := b.String()
	hash := xxhash.Sum64String(normalized)
	return fmt.Sprintf("%016x", hash)
}
