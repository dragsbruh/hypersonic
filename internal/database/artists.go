package database

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"text/template"

	"github.com/dragsbruh/hypersonic/internal/library"
)

func (db *Database) GetArtist(ctx context.Context, artistHash string) (*library.HypersonicArtist, error) {
	row := db.QueryRow(ctx, `
		SELECT hash, name
		FROM artist
		WHERE artist.hash = $1
		LIMIT 1
	`, artistHash,
	)
	artist := DatabaseArtist{}
	if err := row.Scan(&artist.Hash, &artist.Name); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}
	return artist.Hyper(), nil
}

func (db *Database) CreateArtist(ctx context.Context, artist library.HypersonicArtist) error {
	_, err := db.Exec(ctx, `
		INSERT INTO artist (hash, name) VALUES ($1, $2)
		ON CONFLICT (hash) DO NOTHING
	`, artist.Hash, artist.Name)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (db *Database) DeleteArtist(ctx context.Context, artistHash string) error {
	_, err := db.Exec(ctx, `--sql
		DELETE FROM artist
		WHERE hash = $1
	`, artistHash)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (db *Database) RelateTrackArtist(ctx context.Context, trackHash, artistHash string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO track_artist (track_hash, artist_hash)
		VALUES ($1, $2)
		ON CONFLICT (track_hash, artist_hash) DO NOTHING
	`, trackHash, artistHash)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (db *Database) RelateAlbumArtist(ctx context.Context, albumHash, artistHash string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO album_artist (album_hash, artist_hash)
		VALUES ($1, $2)
		ON CONFLICT (album_hash, artist_hash) DO NOTHING
	`, albumHash, artistHash)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (db *Database) GetPaginatedArtists(ctx context.Context, cfg PaginationConfig) (*PaginatedResult[library.HypersonicArtist], error) {
	wr := bytes.NewBufferString("")
	err := template.Must(template.New("").Parse(`
		SELECT
			artist.hash, artist.name
		FROM artist
		
		{{if eq .Column "random" -}}
			ORDER BY hashtext(album.hash || '{{ .Direction }}')
		{{else -}}
			ORDER BY artist.{{ .Column }} {{ .Direction }} 
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

	artists := []library.HypersonicArtist{}
	for rows.Next() {
		dar := DatabaseArtist{}

		if err := rows.Scan(
			&dar.Hash, &dar.Name,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		artists = append(artists, *dar.Hyper())
	}

	totalRow := db.QueryRow(ctx, `SELECT COUNT(*) FROM artist`)
	total := 0
	if err := totalRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("get total rows: %w", err)
	}

	return &PaginatedResult[library.HypersonicArtist]{
		Items:      artists,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(cfg.Limit))),
	}, nil
}
