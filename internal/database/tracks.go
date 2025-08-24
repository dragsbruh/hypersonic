package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"text/template"

	"github.com/dragsbruh/hypersonic/internal/library"
)

func (db *Database) CreateTrack(ctx context.Context, track library.HypersonicTrack) error {
	var albHash *string = nil
	if track.Album != nil {
		albHash = &track.Album.Hash
	}
	_, err := db.Exec(ctx, `
		INSERT INTO track (
			hash, name,
			album_hash, track_number, disc_number,
			duration, year, genre,
			file_path, file_hash, indexed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (hash) DO UPDATE SET
			name = EXCLUDED.name,
			album_hash = EXCLUDED.album_hash,
			track_number = EXCLUDED.track_number,
			disc_number = EXCLUDED.disc_number,
			duration = EXCLUDED.duration,
			year = EXCLUDED.year,
			genre = EXCLUDED.genre,
			file_path = EXCLUDED.file_path,
			file_hash = EXCLUDED.file_hash
	`, track.Hash, track.Name, albHash, track.TrackNumber, track.DiscNumber, track.Duration, track.Year, track.Genre, track.FilePath, track.FileHash, track.IndexedAt)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

// does not include album artists
func (db *Database) GetTrack(ctx context.Context, trackHash string) (*library.HypersonicTrack, error) {
	row := db.QueryRow(ctx, `
		SELECT
			track.hash, track.name,
			track.album_hash, track.track_number, track.disc_number,
			track.duration, track.year, track.genre,
			track.file_path, track.file_hash, track.indexed_at,
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM track
		LEFT JOIN album ON track.album_hash = album.hash
		LEFT JOIN track_artist ON track_artist.track_hash = track.hash
		LEFT JOIN artist ON track_artist.artist_hash = artist.hash
		WHERE track.hash = $1
		GROUP BY track.hash, album.hash
	`, trackHash)
	dt := DatabaseTrack{}
	da := &DatabaseAlbum{}
	artistsJSON := []byte{}
	artists := []library.HypersonicArtist{}

	err := row.Scan(
		&dt.Hash, &dt.Name,
		&dt.AlbumHash, &dt.TrackNumber, &dt.DiscNumber,
		&dt.Duration, &dt.Year, &dt.Genre,
		&dt.FilePath, &dt.FileHash, &dt.IndexedAt,
		&da.Hash, &da.Name, &da.Artwork,
		&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
		&artistsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	var album *library.HypersonicAlbum = nil
	if dt.AlbumHash.Valid {
		album = da.Hyper(nil)
	}

	if err := json.Unmarshal(artistsJSON, &artists); err != nil {
		return nil, fmt.Errorf("unmarshal artists: %w", err)
	}

	hyper := dt.Hyper(album, artists)

	return hyper, nil
}

// does not include album artists
func (db *Database) GetTracksByArtist(ctx context.Context, artistHash string) ([]library.HypersonicTrack, error) {
	rows, err := db.Query(ctx, `
		SELECT
			track.hash, track.name,
			track.album_hash, track.track_number, track.disc_number,
			track.duration, track.year, track.genre,
			track.file_path, track.file_hash, track.indexed_at,
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM track_artist
		JOIN track ON track_artist.track_hash = track.hash
		LEFT JOIN album ON track.album_hash = album.hash
		LEFT JOIN artist ON track_artist.artist_hash = artist.hash
		WHERE track_artist.artist_hash = $1
		GROUP BY track.hash, album.hash
	`, artistHash)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	tracks := []library.HypersonicTrack{}
	for rows.Next() {
		dt := DatabaseTrack{}
		da := &DatabaseAlbum{}
		artistsJSON := []byte{}

		err := rows.Scan(
			&dt.Hash, &dt.Name,
			&dt.AlbumHash, &dt.TrackNumber, &dt.DiscNumber,
			&dt.Duration, &dt.Year, &dt.Genre,
			&dt.FilePath, &dt.FileHash, &dt.IndexedAt,
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		var album *library.HypersonicAlbum = nil
		if dt.AlbumHash.Valid {
			album = da.Hyper(nil)
		}

		artists := []library.HypersonicArtist{}
		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artist: %w", err)
		}

		tracks = append(tracks, *dt.Hyper(album, artists))
	}

	return tracks, nil
}

// does not include album artists
func (db *Database) GetTracksByAlbum(ctx context.Context, albumHash string) ([]library.HypersonicTrack, error) {
	rows, err := db.Query(ctx, `
		SELECT
			track.hash, track.name,
			track.album_hash, track.track_number, track.disc_number,
			track.duration, track.year, track.genre,
			track.file_path, track.file_hash, track.indexed_at,
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM track
		LEFT JOIN album ON album.hash = track.album_hash
		LEFT JOIN track_artist ON track_artist.track_hash = track.hash
		LEFT JOIN artist ON track_artist.artist_hash = artist.hash
		WHERE track.album_hash = $1
		GROUP BY track.hash, album.hash
	`, albumHash)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	tracks := []library.HypersonicTrack{}
	for rows.Next() {
		dt := DatabaseTrack{}
		da := &DatabaseAlbum{}
		artistsJSON := []byte{}

		err := rows.Scan(
			&dt.Hash, &dt.Name,
			&dt.AlbumHash, &dt.TrackNumber, &dt.DiscNumber,
			&dt.Duration, &dt.Year, &dt.Genre,
			&dt.FilePath, &dt.FileHash, &dt.IndexedAt,
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		var album *library.HypersonicAlbum = nil
		if dt.AlbumHash.Valid {
			album = da.Hyper(nil)
		}

		artists := []library.HypersonicArtist{}
		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artists: %w", err)
		}

		tracks = append(tracks, *dt.Hyper(album, artists))
	}

	return tracks, nil
}

func (db *Database) DeleteTrack(ctx context.Context, trackHash string) error {
	_, err := db.Exec(ctx, `--sql
		DELETE FROM track
		WHERE hash = $1
	`, trackHash)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (db *Database) GetFileHashes(ctx context.Context) (map[string]string, error) {
	rows, err := db.Query(ctx, `SELECT hash, file_hash FROM track`)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	defer rows.Close()

	hashes := map[string]string{}
	for rows.Next() {
		h := TrackHashes{}
		if err := rows.Scan(&h.TrackHash, &h.FileHash); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		hashes[h.FileHash] = h.TrackHash
	}

	return hashes, nil
}

func (db *Database) GetPaginatedTracks(ctx context.Context, cfg PaginationConfig) (*PaginatedResult[library.HypersonicTrack], error) {
	wr := bytes.NewBufferString("")
	err := template.Must(template.New("").Parse(`
		SELECT
			track.hash, track.name,
			track.album_hash, track.track_number, track.disc_number,
			track.duration, track.year, track.genre,
			track.file_path, track.file_hash, track.indexed_at,
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM track
		LEFT JOIN album ON album.hash = track.album_hash
		LEFT JOIN track_artist ON track_artist.track_hash = track.hash
		LEFT JOIN artist ON track_artist.artist_hash = artist.hash

		GROUP BY track.hash, album.hash
		{{if eq .Column "random" -}}
			ORDER BY hashtext(track.hash || '{{ .Direction }}')
		{{else -}}
			ORDER BY track.{{ .Column }} {{ .Direction }} 
		{{end}}
		LIMIT $1 OFFSET $2
	`)).Execute(wr, cfg)
	if err != nil {
		return nil, fmt.Errorf("exec template: %w", err)
	}
	sql := wr.String()

	rows, err := db.Query(ctx, sql, cfg.Limit, (cfg.Page-1)*cfg.Limit)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	tracks := []library.HypersonicTrack{}
	for rows.Next() {
		dt := DatabaseTrack{}
		da := DatabaseAlbum{}
		artistsJSON := []byte{}

		err := rows.Scan(&dt.Hash, &dt.Name,
			&dt.AlbumHash, &dt.TrackNumber, &dt.DiscNumber,
			&dt.Duration, &dt.Year, &dt.Genre,
			&dt.FilePath, &dt.FileHash, &dt.IndexedAt,
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		var album *library.HypersonicAlbum = nil
		if dt.AlbumHash.Valid {
			album = da.Hyper(nil)
		}

		artists := []library.HypersonicArtist{}
		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artist: %w", err)
		}

		tracks = append(tracks, *dt.Hyper(album, artists))
	}

	totalRow := db.QueryRow(ctx, `SELECT COUNT(*) FROM track`)
	total := 0
	if err := totalRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("get total rows: %w", err)
	}

	return &PaginatedResult[library.HypersonicTrack]{
		Items:      tracks,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(cfg.Limit))),
	}, nil
}

func (db *Database) GetTracksByGenre(ctx context.Context, genre string) ([]library.HypersonicTrack, error) {
	rows, err := db.Query(ctx, `
		SELECT
			track.hash, track.name,
			track.album_hash, track.track_number, track.disc_number,
			track.duration, track.year, track.genre,
			track.file_path, track.file_hash, track.indexed_at,
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM track
		LEFT JOIN album ON album.hash = track.album_hash
		LEFT JOIN track_artist ON track_artist.track_hash = track.hash
		LEFT JOIN artist ON track_artist.artist_hash = artist.hash
		WHERE track.genre = $1
		GROUP BY track.hash, album.hash
	`, genre)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	tracks := []library.HypersonicTrack{}
	for rows.Next() {
		dt := DatabaseTrack{}
		da := &DatabaseAlbum{}
		artistsJSON := []byte{}

		err := rows.Scan(
			&dt.Hash, &dt.Name,
			&dt.AlbumHash, &dt.TrackNumber, &dt.DiscNumber,
			&dt.Duration, &dt.Year, &dt.Genre,
			&dt.FilePath, &dt.FileHash, &dt.IndexedAt,
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		var album *library.HypersonicAlbum = nil
		if dt.AlbumHash.Valid {
			album = da.Hyper(nil)
		}

		artists := []library.HypersonicArtist{}
		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artists: %w", err)
		}

		tracks = append(tracks, *dt.Hyper(album, artists))
	}

	return tracks, nil
}
