package scan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/dhowden/tag"
	"github.com/dragsbruh/hypersonic/internal/media"
	"github.com/dragsbruh/hypersonic/internal/metadata"
	"github.com/sirupsen/logrus"
)

var MusicExts = []string{
	".flac", ".mp3", ".ogg", ".wav", ".m4a", ".aac",
	".alac", ".aiff", ".opus", ".wma",
}

func ScanAudioMetadata(audioFilePath string) (*metadata.HypersonicTrack, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	tags, err := tag.ReadFrom(file)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			logrus.Debugf("failed to read metadata for %s, returning default metadata", audioFilePath)
			return &metadata.HypersonicTrack{
				Hash:  metadata.CalculateHash(file.Name()),
				Title: file.Name(),
			}, nil
		}
		return nil, fmt.Errorf("read tags: %w", err)
	}

	absPath, err := filepath.Abs(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("abs path: %w", err)
	}

	discNumber, totalDiscs := tags.Disc()
	trackNumber, totalTracks := tags.Track()

	album := metadata.HypersonicAlbum{
		Title:       tags.Album(),
		Artists:     parseArtists(tags.AlbumArtist()),
		TotalDiscs:  totalDiscs,
		TotalTracks: totalTracks,
	}
	album.Hash = metadata.CalculateHash(album.Title, album.Artists[0].Title)

	picture := tags.Picture()
	if picture != nil {
		if err := media.StoreAlbumArt(album.Hash, picture.Data); err != nil {
			return nil, err
		}
	}

	track := metadata.HypersonicTrack{
		Title:       tags.Title(),
		Artists:     parseArtists(tags.Artist()),
		Album:       &album,
		DiscNumber:  discNumber,
		TrackNumber: trackNumber,
		Year:        tags.Year(),
		Genre:       tags.Genre(),
		Composer:    tags.Composer(),
		FilePath:    absPath,
	}
	track.Hash = metadata.CalculateHash(track.Title, track.Artists[0].Title, track.Album.Title)

	return &track, nil
}

func RecursiveScanDirectory(ctx context.Context, dir string, cb func(int, string, string) error) error {
	count := 0

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

outer:
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			break outer
		default:
		}
		fullpath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err := RecursiveScanDirectory(ctx, fullpath, func(_ int, fpath, filehash string) error {
				count += 1
				return cb(count, fpath, filehash)
			})
			if err != nil {
				return fmt.Errorf("scan subdir: %w", err)
			}
		} else if slices.Contains(MusicExts, filepath.Ext(fullpath)) {
			digest := xxhash.New()

			file, err := os.Open(fullpath)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer file.Close()

			io.Copy(digest, file)

			count += 1
			if err := cb(count, fullpath, fmt.Sprintf("%016x", digest.Sum64())); err != nil {
				return fmt.Errorf("cb error: %w", err)
			}
		}
	}

	return nil
}

func parseArtists(raw string) []metadata.HypersonicArtist {
	artists := []metadata.HypersonicArtist{}

	var iter iter.Seq[string]
	if strings.Contains(raw, ";") {
		iter = strings.SplitSeq(raw, ";")
	} else {
		iter = strings.SplitSeq(raw, ",")
	}

	for artist := range iter {
		artists = append(artists, metadata.HypersonicArtist{
			Hash:  metadata.CalculateHash(artist),
			Title: strings.TrimSpace(artist),
		})
	}
	return artists
}
