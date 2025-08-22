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

// includes artists
func (db *Database) GetAlbum(ctx context.Context, albumHash string) (*library.HypersonicAlbum, error) {
	row := db.QueryRow(ctx, `
		SELECT
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM album
		LEFT JOIN album_artist ON album_artist.album_hash = album.hash
		LEFT JOIN artist ON album_artist.artist_hash = artist.hash
		WHERE album.hash = $1
		GROUP BY album.hash
	`, albumHash)

	album := DatabaseAlbum{}
	artistsJSON := []byte{}

	if err := row.Scan(
		&album.Hash, &album.Name, &album.Artwork,
		&album.TotalTracks, &album.TotalDiscs, &album.IndexedAt,
		&artistsJSON,
	); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	artists := []library.HypersonicArtist{}
	if err := json.Unmarshal(artistsJSON, &artists); err != nil {
		return nil, fmt.Errorf("unmarshal artist: %w", err)
	}

	return album.Hyper(artists), nil
}

// does not insert artists
func (db *Database) CreateAlbum(ctx context.Context, album library.HypersonicAlbum) error {
	_, err := db.Exec(ctx, `
		INSERT INTO
			album (hash, name, artwork, total_tracks, total_discs, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (hash) DO UPDATE SET
			artwork = EXCLUDED.artwork,
			total_tracks = EXCLUDED.total_tracks,
			total_discs = EXCLUDED.total_discs
	`, album.Hash, album.Name, album.Artwork, album.TotalTracks, album.TotalDiscs, album.IndexedAt)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (db *Database) DeleteAlbum(ctx context.Context, albumHash string) error {
	_, err := db.Exec(ctx, `--sql
		DELETE FROM album
		WHERE hash = $1
	`, albumHash)

	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

// includes artists
func (db *Database) GetAlbumsByArtist(ctx context.Context, artistHash string) ([]library.HypersonicAlbum, error) {
	rows, err := db.Query(ctx, `
		SELECT
		    album.hash, album.name, album.artwork,
		    album.total_tracks, album.total_discs, album.indexed_at,
		    COALESCE(json_agg(json_build_object(
		        'hash', artist.hash,
		        'name', artist.name
		    )) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM album_artist
		JOIN album ON album_artist.album_hash = album.hash
		LEFT JOIN album_artist ar_a ON album.hash = ar_a.album_hash
		LEFT JOIN artist ON ar_a.artist_hash = artist.hash
		WHERE album_artist.artist_hash = $1
		GROUP BY album.hash
	`, artistHash)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	albums := []library.HypersonicAlbum{}
	for rows.Next() {
		da := DatabaseAlbum{}
		artists := []library.HypersonicArtist{}
		artistsJSON := []byte{}

		err := rows.Scan(
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artist: %w", err)
		}

		albums = append(albums, *da.Hyper(artists))
	}

	return albums, nil
}

func (db *Database) GetPaginatedAlbums(ctx context.Context, cfg PaginationConfig) (*PaginatedResult[library.HypersonicAlbum], error) {
	wr := bytes.NewBufferString("")
	err := template.Must(template.New("").Parse(`
		SELECT
			album.hash, album.name, album.artwork,
			album.total_tracks, album.total_discs, album.indexed_at,
			COALESCE(json_agg(json_build_object(
				'hash', artist.hash,
				'name', artist.name
			)) FILTER (WHERE artist.hash IS NOT NULL), '[]'::json) AS artists
		FROM album
		LEFT JOIN album_artist ON album_artist.album_hash = album.hash
		LEFT JOIN artist ON album_artist.artist_hash = artist.hash

		GROUP BY album.hash
		{{if eq .Column "random" -}}
			ORDER BY hashtext(album.hash || '{{ .Direction }}')
		{{else -}}
			ORDER BY album.{{ .Column }} {{ .Direction }} 
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

	albums := []library.HypersonicAlbum{}
	for rows.Next() {
		da := DatabaseAlbum{}
		artistsJSON := []byte{}

		err := rows.Scan(
			&da.Hash, &da.Name, &da.Artwork,
			&da.TotalTracks, &da.TotalDiscs, &da.IndexedAt,
			&artistsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		artists := []library.HypersonicArtist{}
		if err := json.Unmarshal(artistsJSON, &artists); err != nil {
			return nil, fmt.Errorf("unmarshal artist: %w", err)
		}

		albums = append(albums, *da.Hyper(artists))
	}

	totalRow := db.QueryRow(ctx, `SELECT COUNT(*) FROM album`)
	total := 0
	if err := totalRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("get total rows: %w", err)
	}

	return &PaginatedResult[library.HypersonicAlbum]{
		Items:      albums,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(cfg.Limit))),
	}, nil
}
