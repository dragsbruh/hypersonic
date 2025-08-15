package metadata

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/mozillazg/go-unidecode"
)

type HypersonicArtist struct {
	Hash  string
	Title string
}

type HypersonicAlbum struct {
	Hash        string
	Title       string
	Artists     []HypersonicArtist
	TotalDiscs  int
	TotalTracks int
}

type HypersonicTrack struct {
	Hash        string // track hash, not hash of file
	Title       string
	Artists     []HypersonicArtist
	Album       *HypersonicAlbum
	DiscNumber  int
	TrackNumber int
	Year        int
	Genre       string
	Composer    string
	FilePath    string
	FileHash    string
}

func normalize(s string) string {
	s = unidecode.Unidecode(strings.ToLower(strings.TrimSpace(s)))
	sb := strings.Builder{}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func CalculateHash(args ...string) string {
	combined := ""
	for _, a := range args {
		combined += normalize(a)
	}
	digest := xxhash.New()
	digest.WriteString(combined)
	return fmt.Sprintf("%016x", digest.Sum64())
}
