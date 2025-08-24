package library

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/mozillazg/go-unidecode"
)

type HypersonicArtist struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
}

func (a HypersonicArtist) FreshHash() string {
	return GetHashOf(a.Name)
}

type HypersonicAlbum struct {
	Hash string `json:"hash"`
	Name string `json:"name"`

	Artists []HypersonicArtist `json:"artists"`

	Artwork     bool `json:"artwork"`
	TotalTracks *int `json:"total_tracks"`
	TotalDiscs  *int `json:"total_discs"`

	IndexedAt time.Time `json:"indexed_at"`
}

func (a HypersonicAlbum) FreshHash() string {
	firstArtist := ""
	if len(a.Artists) > 0 {
		firstArtist = a.Artists[0].Name
	}
	return GetHashOf(a.Name, firstArtist)
}

type HypersonicTrack struct {
	Hash string `json:"hash"`
	Name string `json:"name"`

	Artists []HypersonicArtist `json:"artists"`

	Album       *HypersonicAlbum `json:"album"`
	TrackNumber int              `json:"track_number"`
	DiscNumber  int              `json:"disc_number"`

	Duration int     `json:"duration"`
	Year     *int    `json:"year"`
	Genre    *string `json:"genre"`

	FilePath string `json:"file_path"`
	FileHash string `json:"file_hash"`

	IndexedAt time.Time `json:"indexed_at"`
}

func (t HypersonicTrack) FreshHash() string {
	firstArtist := ""
	if len(t.Artists) > 0 {
		firstArtist = t.Artists[0].Name
	}
	albumName := ""
	if t.Album != nil {
		albumName = t.Album.Name
	}
	return GetHashOf(t.Name, albumName, firstArtist)
}

func GetHashOf(args ...string) string {
	var b strings.Builder
	for _, arg := range args {
		for _, r := range unidecode.Unidecode(arg) {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(unicode.ToLower(r))
			}
		}
	}
	return strconv.FormatUint(xxhash.Sum64String(b.String()), 16)
}
