package library

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/mozillazg/go-unidecode"
)

type HypersonicArtist struct {
	Hash string
	Name string
}

type HypersonicAlbum struct {
	Hash string
	Name string

	Artists []HypersonicArtist

	TotalTracks int
	TotalDiscs  int
}

type HypersonicTrack struct {
	Hash string
	Name string

	Artists []HypersonicArtist

	Album       *HypersonicAlbum
	DiscNumber  int
	TrackNumber int
	Year        int
	Genre       string

	FilePath string
	FileHash string
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
