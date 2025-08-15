package media

import (
	"context"
	"fmt"

	"github.com/dragsbruh/hypersonic/internal/database/queries"
	"github.com/dragsbruh/hypersonic/internal/metadata"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateArtist(ctx context.Context, q *queries.Queries, artist metadata.HypersonicArtist) error {
	err := q.CreateArtist(ctx, queries.CreateArtistParams{
		ID:    artist.Hash,
		Title: artist.Title,
	})
	if err != nil {
		return fmt.Errorf("create artist db: %w", err)
	}
	return nil
}

func CreateAlbum(ctx context.Context, q *queries.Queries, album metadata.HypersonicAlbum) error {
	err := q.CreateAlbum(ctx, queries.CreateAlbumParams{
		ID:          album.Hash,
		Title:       album.Title,
		TotalDiscs:  int32(album.TotalDiscs),
		TotalTracks: int32(album.TotalTracks),
	})
	if err != nil {
		return fmt.Errorf("create album db: %w", err)
	}
	for _, artist := range album.Artists {
		if err := CreateArtist(ctx, q, artist); err != nil {
			return fmt.Errorf("create artist: %w", err)
		}
		if err := q.MarkAlbumArtist(ctx, queries.MarkAlbumArtistParams{
			ArtistID: artist.Hash,
			AlbumID:  album.Hash,
		}); err != nil {
			return fmt.Errorf("mark artist: %w", err)
		}
	}
	return nil
}

func CreateTrack(ctx context.Context, q *queries.Queries, track metadata.HypersonicTrack) error {
	if track.Album != nil {
		if err := CreateAlbum(ctx, q, *track.Album); err != nil {
			return fmt.Errorf("create album: %w", err)
		}
	}

	albumID := pgtype.Text{Valid: false}
	if track.Album != nil {
		albumID.Valid = true
		albumID.String = track.Album.Hash
	}

	err := q.CreateTrack(ctx, queries.CreateTrackParams{
		ID:          track.Hash,
		Title:       track.Title,
		AlbumID:     albumID,
		DiscNumber:  int32(track.DiscNumber),
		TrackNumber: int32(track.TrackNumber),
		Year:        int32(track.Year),
		Genre:       track.Genre,
		Composer:    track.Composer,
		FilePath:    track.FilePath,
		FileHash:    track.FileHash,
	})
	if err != nil {
		return fmt.Errorf("create track db: %w", err)
	}

	for _, artist := range track.Artists {
		if err := CreateArtist(ctx, q, artist); err != nil {
			return fmt.Errorf("create artist: %w", err)
		}
		if err := q.MarkTrackArtist(ctx, queries.MarkTrackArtistParams{
			ArtistID: artist.Hash,
			TrackID:  track.Hash,
		}); err != nil {
			return fmt.Errorf("mark artist: %w", err)
		}
	}
	return nil
}

type ComparableTrack struct {
	Hash     string
	FileHash string
	FilePath string
}

func GetComparableTracks(ctx context.Context, q *queries.Queries) ([]ComparableTrack, error) {
	tracks, err := q.GetTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tracks db: %w", err)
	}

	comparable := []ComparableTrack{}
	for _, t := range tracks {
		comparable = append(comparable, ComparableTrack{
			Hash:     t.ID,
			FileHash: t.FileHash,
			FilePath: t.FilePath,
		})
	}

	return comparable, nil
}

func DeleteTrack(ctx context.Context, q *queries.Queries, trackHash string) error {
	err := q.DeleteTrack(ctx, trackHash)
	if err != nil {
		return fmt.Errorf("delete track db: %w", err)
	}
	return nil
}
