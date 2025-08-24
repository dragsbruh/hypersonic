package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/library"
	"github.com/sirupsen/logrus"
)

var AudioExtensions = []string{
	".ogg", ".mp3", ".wav", ".m4a", ".webm", ".flac", ".opus",
}

type FileHashPair struct {
	FilePath string
	FileHash string
}

func ScanDirectory(ctx context.Context, dir string) ([]string, error) {
	files := []string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
			fullpath := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				subItems, err := ScanDirectory(ctx, fullpath)
				if err != nil {
					return nil, err
				}
				files = append(files, subItems...)
			} else {
				if slices.Contains(AudioExtensions, filepath.Ext(fullpath)) {
					files = append(files, fullpath)
				}
			}
		}
	}

	return files, nil
}

type ScanStatus struct {
	Status  string `json:"status"` // "add" "remove" "skip" "error" or "info"
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
	Name    string `json:"title,omitempty"`
}

func ScanAndUpdate(ctx context.Context, db *database.Database, scb func(ScanStatus)) error {
	scb(ScanStatus{
		Status:  "info",
		Message: "scanning directory for files",
	})
	trackPaths, err := ScanDirectory(ctx, config.MusicRoot)
	if err != nil {
		logrus.Fatalf("error scanning music root: %v", err)
	}

	scb(ScanStatus{
		Status:  "info",
		Message: "get database state",
	})
	indexedFileHashes, err := db.GetFileHashes(ctx)
	if err != nil {
		return fmt.Errorf("get track hashes: %w", err)
	}

	scb(ScanStatus{
		Status:  "info",
		Message: "scanning metadata and indexing tracks",
	})
	seenFileHashes := sync.Map{}
	completedItems := sync.Map{}

	scb(ScanStatus{
		Status:  "info",
		Message: "create worker pool",
	})
	err = NewWorkerPool(trackPaths, config.Concurrency, func(ctx context.Context, trackPath string) (*library.HypersonicTrack, error) {
		hash, err := GetFileHash(trackPath)
		if err != nil {
			return nil, fmt.Errorf("get hash: %w", err)
		}

		trackHash, ok := indexedFileHashes[hash]
		if ok {
			seenFileHashes.Store(hash, trackHash)
			scb(ScanStatus{
				Status: "skip",
				Path:   trackPath,
				Name:   trackHash,
			})
			return nil, nil
		}

		track, err := ScanFile(trackPath, hash)
		if err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		seenFileHashes.Store(hash, track.Hash)

		if track.Album != nil {
			_, ok := completedItems.LoadOrStore(fmt.Sprintf("artwork_%s", track.Album.Hash), struct{}{})
			if !ok {
				if track.Album.Artwork {
					if err := ExtractArtwork(track.Album.Hash, trackPath); err != nil {
						return nil, fmt.Errorf("extract album artwork: %w", err)
					}
				}
			}
		}

		return track, nil
	}).Run(ctx, func(ctx context.Context, ht *library.HypersonicTrack) error {
		if ht == nil {
			return nil
		}

		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("create tx: %w", err)
		}
		defer tx.Rollback(ctx)

		if ht.Album != nil {
			_, ok := completedItems.LoadOrStore(fmt.Sprintf("album_%s", ht.Album.Hash), struct{}{})
			if !ok {
				if err := tx.CreateAlbum(ctx, *ht.Album); err != nil {
					return fmt.Errorf("create album: %w", err)
				}
			}
		}

		if err := tx.CreateTrack(ctx, *ht); err != nil {
			return fmt.Errorf("create track: %w", err)
		}

		for _, artist := range ht.Artists {
			_, ok := completedItems.LoadOrStore(fmt.Sprintf("artist_%s", artist.Hash), struct{}{})
			if !ok {
				if err := tx.CreateArtist(ctx, artist); err != nil {
					return fmt.Errorf("create artist: %w", err)
				}
			}
			if err := tx.RelateTrackArtist(ctx, ht.Hash, artist.Hash); err != nil {
				return fmt.Errorf("relate track artist: %w", err)
			}
		}

		if ht.Album != nil {
			for _, artist := range ht.Album.Artists {
				_, ok := completedItems.LoadOrStore(fmt.Sprintf("artist_%s", artist.Hash), struct{}{})
				if !ok {
					if err := tx.CreateArtist(ctx, artist); err != nil {
						return fmt.Errorf("create artist: %w", err)
					}
				}
				if err := tx.RelateAlbumArtist(ctx, ht.Album.Hash, artist.Hash); err != nil {
					return fmt.Errorf("relate album artist: %w", err)
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}

		scb(ScanStatus{
			Status: "add",
			Path:   ht.FilePath,
			Name:   ht.Name,
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("run pool: %w", err)
	}

	scb(ScanStatus{
		Status:  "info",
		Message: "primary scan complete, deleting missing tracks",
	})
	for fhash, trackHash := range indexedFileHashes {
		_, ok := seenFileHashes.Load(fhash)
		if !ok {
			scb(ScanStatus{
				Status: "delete",
				Name:   trackHash,
			})
			if err := db.DeleteTrack(ctx, trackHash); err != nil {
				return fmt.Errorf("delete track: %v", err)
			}
		}
	}

	scb(ScanStatus{
		Status:  "info",
		Message: "complete",
	})
	return nil
}
