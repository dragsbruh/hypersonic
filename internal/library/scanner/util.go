package scanner

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/dragsbruh/hypersonic/internal/library"
)

var yearRegex = regexp.MustCompile(`\d{4}`)

func ParseYear(dateString string) (int, bool) {
	if dateString == "" {
		return 0, false
	}
	str := yearRegex.FindString(dateString)
	if str == "" {
		return 0, false
	}
	year, err := strconv.Atoi(str)
	if err != nil {
		return 0, false
	}
	return year, true
}

func ParseArtists(artistsString string) []library.HypersonicArtist {
	parsedArtists := []library.HypersonicArtist{}
	iter := strings.SplitSeq(artistsString, ";")
	if !strings.Contains(artistsString, ";") {
		iter = strings.SplitSeq(artistsString, ",")
	}
	for artistName := range iter {
		artistName = strings.TrimSpace(artistName)
		if artistName == "" {
			continue
		}
		artist := library.HypersonicArtist{
			Name: artistName,
		}
		artist.Hash = artist.FreshHash()
		parsedArtists = append(parsedArtists, artist)
	}
	return parsedArtists
}

func GetFileHash(trackPath string) (string, error) {
	file, err := os.Open(trackPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	digest := xxhash.New()
	io.Copy(digest, file)

	return strconv.FormatUint(digest.Sum64(), 16), nil
}
