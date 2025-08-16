package library

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/dhowden/tag"
	"github.com/dragsbruh/hypersonic/internal/storage"
)

func GetTags(trackPath string) (tag.Metadata, error) {
	file, err := os.Open(trackPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()
	return tag.ReadFrom(file)
}

func ParseTrackMetadata(trackPath, fileHash string) (*HypersonicTrack, error) {
	tags, err := GetTags(trackPath)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	discNum, totalDiscs := tags.Disc()
	trackNum, totalTracks := tags.Track()
	picture := tags.Picture()

	var album *HypersonicAlbum = nil
	if tags.Album() != "" {
		album = &HypersonicAlbum{
			Name:        tags.Album(),
			Artists:     parseArtists(tags.AlbumArtist()),
			TotalDiscs:  totalDiscs,
			TotalTracks: totalTracks,
		}
		firstArtist := ""
		if len(album.Artists) > 0 {
			firstArtist = album.Artists[0].Name
		}
		album.Hash = CalculateHash(album.Name, firstArtist)
	}

	track := HypersonicTrack{
		Name:        tags.Title(),
		Artists:     parseArtists(tags.AlbumArtist()),
		Album:       album,
		DiscNumber:  discNum,
		TrackNumber: trackNum,
		Year:        tags.Year(),
		Genre:       tags.Genre(),
		FilePath:    trackPath,
		FileHash:    fileHash,
	}
	firstArtist := ""
	if len(track.Artists) > 0 {
		firstArtist = track.Artists[0].Name
	}
	track.Hash = CalculateHash(track.Name, firstArtist)

	if picture != nil {
		storage.StoreCoverArt(bytes.NewBuffer(picture.Data), album.Hash)
	}

	return &track, nil
}

func parseArtists(str string) []HypersonicArtist {
	artists := []HypersonicArtist{}

	iter := strings.SplitSeq(str, ";")
	if !strings.Contains(str, ";") {
		iter = strings.SplitSeq(str, ",")
	}

	for name := range iter {
		name = strings.TrimSpace(name)
		artists = append(artists, HypersonicArtist{
			Hash: CalculateHash(name),
			Name: name,
		})
	}

	return artists
}

func GetFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	digest := xxhash.New()
	io.Copy(digest, file)
	return fmt.Sprintf("%016x", digest.Sum64()), nil
}
