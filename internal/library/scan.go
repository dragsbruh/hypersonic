package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// recursively scans directory and calls callback with paths of music tracks
// aborts if callback errors or context cancels
func ScanDirectory(ctx context.Context, directory string, callback func(scanned int, trackPath string) error) error {
	count := 0

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ErrScanAborted
		default:
		}

		entryPath := filepath.Join(directory, entry.Name())
		if entry.IsDir() {
			err := ScanDirectory(ctx, entryPath, func(scanned int, trackPath string) error {
				count += 1
				return callback(count, trackPath)
			})
			if err != nil {
				return fmt.Errorf("subdir %s: %w", entry.Name(), err)
			}
		} else {
			count += 1
			if err := callback(count, entryPath); err != nil {
				return fmt.Errorf("callback: %w", err)
			}
		}
	}

	return nil
}

var ErrScanAborted = errors.New("scan aborted (context cancelled)")
