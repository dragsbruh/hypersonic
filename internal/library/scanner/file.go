package scanner

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

	tags := result.Format.Tags
	if tags == nil {
		if audioStream.Tags != nil {
			tags = audioStream.Tags
		} else if videoStream.Tags != nil {
			tags = videoStream.Tags // stupid
		} else {
			tags = &ProbeTags{}
		}
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

	if tags.Album != "" {
		album := library.HypersonicAlbum{
			Name:      tags.Album,
			IndexedAt: time.Now(),
			Artwork:   videoStream != nil,
		}
		if tags.AlbumArtist != "" {
			album.Artists = ParseArtists(tags.AlbumArtist)
		}
		if tags.TrackTotal != "" {
			i, err := strconv.Atoi(tags.TrackTotal)
			if err != nil {
				return nil, fmt.Errorf("parse total tracks: %w", err)
			}
			album.TotalTracks = &i
		}
		if tags.DiscTotal != "" {
			i, err := strconv.Atoi(tags.DiscTotal)
			if err != nil {
				return nil, fmt.Errorf("parse total discs: %w", err)
			}
			album.TotalDiscs = &i
		}

		album.Hash = album.FreshHash()
		track.Album = &album
	}

	if tags.Title != "" {
		track.Name = tags.Title
	}
	if tags.Track != "" {
		if strings.Contains(tags.Track, "/") {
			count, total, err := parseCounts(tags.Track)
			if err != nil {
				return nil, fmt.Errorf("parse track counts: %w", err)
			}
			track.TrackNumber = count
			if track.Album != nil {
				track.Album.TotalTracks = &total
			}
		} else {
			track.TrackNumber, err = strconv.Atoi(tags.Track)
			if err != nil {
				return nil, fmt.Errorf("parse simple track number: %w", err)
			}
		}

	}
	if tags.Disc != "" {
		if strings.Contains(tags.Disc, "/") {
			count, total, err := parseCounts(tags.Disc)
			if err != nil {
				return nil, fmt.Errorf("parse disc counts: %w", err)
			}
			track.DiscNumber = count
			if track.Album != nil {
				track.Album.TotalDiscs = &total
			}
		} else {
			track.DiscNumber, err = strconv.Atoi(tags.Disc)
			if err != nil {
				return nil, fmt.Errorf("parse simple disc number: %w", err)
			}
		}
	}
	if tags.Artist != "" {
		track.Artists = ParseArtists(tags.Artist)
	}
	if tags.Date != "" {
		y, ok := ParseYear(tags.Date)
		if ok {
			track.Year = &y
		}
	}
	if tags.Genre != "" {
		track.Genre = &tags.Genre
	}

	track.Hash = track.FreshHash()

	return &track, nil
}

func parseCounts(s string) (int, int, error) {
	parts := strings.Split(s, "/")

	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse count: %w", err)
	}

	total, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse total: %w", err)
	}

	return count, total, nil
}
