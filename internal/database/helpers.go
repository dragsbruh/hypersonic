package database

import (
	"context"
	"fmt"

	"github.com/dragsbruh/hypersonic/internal/database/query"
	"github.com/dragsbruh/hypersonic/internal/library"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// returns map where key is file hash and value is track hash for every track in database
func GetExistingHashes(ctx context.Context, queries *PGQueries) (map[string]string, error) {
	rows, err := queries.GetTrackHashes(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	var existing = map[string]string{}
	for _, row := range rows {
		existing[row.FileHash] = row.Hash
	}

	return existing, nil
}

// creates transaction for track
// TODO: probably create tx for album and also dont insert if album exists
func CreateTrack(ctx context.Context, queries *PGQueries, track *library.HypersonicTrack) error {
	tx, q, err := queries.Tx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	albumHash := pgtype.Text{Valid: false}
	if track.Album != nil {
		albumHash.Valid = true
		albumHash.String = track.Album.Hash
		err = q.CreateAlbum(ctx, query.CreateAlbumParams{
			Hash:        track.Album.Hash,
			Name:        track.Album.Name,
			TotalTracks: int32(track.Album.TotalTracks),
			TotalDiscs:  int32(track.Album.TotalDiscs),
		})
		if err != nil {
			return fmt.Errorf("create album: %w", err)
		}

		for _, artist := range track.Album.Artists {
			err := q.CreateArtist(ctx, query.CreateArtistParams{
				Hash: artist.Hash,
				Name: artist.Name,
			})
			if err != nil {
				return fmt.Errorf("create artist: %w", err)
			}

			err = q.MarkAlbumArtist(ctx, query.MarkAlbumArtistParams{
				AlbumHash:  track.Album.Hash,
				ArtistHash: artist.Hash,
			})
			if err != nil {
				return fmt.Errorf("mark album artist: %w", err)
			}
		}
	}

	err = q.CreateTrack(ctx, query.CreateTrackParams{
		Hash:        track.Hash,
		Name:        track.Name,
		AlbumHash:   albumHash,
		DiscNumber:  int32(track.DiscNumber),
		TrackNumber: int32(track.TrackNumber),
		Year:        int32(track.Year),
		Genre:       track.Genre,
		FilePath:    track.FilePath,
		FileHash:    track.FileHash,
	})
	if err != nil {
		return fmt.Errorf("create track db: %w", err)
	}

	for _, artist := range track.Artists {
		err := q.CreateArtist(ctx, query.CreateArtistParams{
			Hash: artist.Hash,
			Name: artist.Name,
		})
		if err != nil {
			return fmt.Errorf("create artist: %w", err)
		}

		err = q.MarkTrackArtist(ctx, query.MarkTrackArtistParams{
			TrackHash:  track.Hash,
			ArtistHash: artist.Hash,
		})
		if err != nil {
			return fmt.Errorf("mark track artist: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
