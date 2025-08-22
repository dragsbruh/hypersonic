package scanner

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/dragsbruh/hypersonic/internal/library"
)

func ScanFile(trackPath, knownFileHash string) (*library.HypersonicTrack, error) {
	if knownFileHash == "" {
		var err error
		knownFileHash, err = GetFileHash(trackPath)
		if err != nil {
			return nil, fmt.Errorf("get file hash: %w", err)
		}
	}

	result, err := GetProbe(trackPath)
	if err != nil {
		return nil, fmt.Errorf("probe: %w", err)
	}

	var audioStream, videoStream *ProbeStream

	for _, stream := range result.Streams {
		if stream.CodecType == "audio" && audioStream == nil {
			audioStream = &stream
		} else if stream.CodecType == "video" && stream.Disposition.AttachedPic == 1 {
			videoStream = &stream
		}
	}

	if audioStream == nil {
		return nil, fmt.Errorf("missing audio stream")
	}

	duration, err := strconv.ParseFloat(audioStream.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("parse duration: %w", err)
	}

	track := library.HypersonicTrack{
		Name:        trackPath,
		Duration:    int(math.Round(duration)),
		TrackNumber: 1,
		DiscNumber:  1,
		FilePath:    trackPath,
		FileHash:    knownFileHash,
		IndexedAt:   time.Now(),
	}
	if audioStream.Tags.Title != "" {
		track.Name = audioStream.Tags.Title
	}
	if audioStream.Tags.Track != "" {
		track.TrackNumber, err = strconv.Atoi(audioStream.Tags.Track)
		if err != nil {
			return nil, fmt.Errorf("parse track number: %w", err)
		}
	}
	if audioStream.Tags.Disc != "" {
		track.DiscNumber, err = strconv.Atoi(audioStream.Tags.Disc)
		if err != nil {
			return nil, fmt.Errorf("parse disc number: %w", err)
		}
	}
	if audioStream.Tags.Artist != "" {
		track.Artists = ParseArtists(audioStream.Tags.Artist)
	}
	if audioStream.Tags.Date != "" {
		y, ok := ParseYear(audioStream.Tags.Date)
		if ok {
			track.Year = &y
		}
	}
	if audioStream.Tags.Genre != "" {
		track.Genre = &audioStream.Tags.Genre
	}

	if audioStream.Tags.Album != "" {
		album := library.HypersonicAlbum{
			Name:      audioStream.Tags.Album,
			IndexedAt: time.Now(),
			Artwork:   videoStream != nil,
		}
		if audioStream.Tags.AlbumArtist != "" {
			album.Artists = ParseArtists(audioStream.Tags.AlbumArtist)
		}
		if audioStream.Tags.TrackTotal != "" {
			i, err := strconv.Atoi(audioStream.Tags.TrackTotal)
			if err != nil {
				return nil, fmt.Errorf("parse total tracks: %w", err)
			}
			album.TotalTracks = &i
		}
		if audioStream.Tags.DiscTotal != "" {
			i, err := strconv.Atoi(audioStream.Tags.DiscTotal)
			if err != nil {
				return nil, fmt.Errorf("parse total discs: %w", err)
			}
			album.TotalDiscs = &i
		}

		album.Hash = album.FreshHash()
		track.Album = &album

	}
	track.Hash = track.FreshHash()

	return &track, nil
}
