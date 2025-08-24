package database

import (
	"database/sql"
	"time"

	"github.com/dragsbruh/hypersonic/internal/library"
)

type DatabaseUser struct {
	ID        string       `json:"id"`
	Username  string       `json:"username"`
	Password  string       `json:"password,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	LoginAt   sql.NullTime `json:"loginAt"`
}

type DatabaseArtist struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
}

func (da DatabaseArtist) Hyper() *library.HypersonicArtist {
	return &library.HypersonicArtist{
		Hash: da.Hash,
		Name: da.Name,
	}
}

type DatabaseAlbum struct {
	Hash        sql.NullString
	Name        sql.NullString
	Artwork     sql.NullBool
	TotalTracks sql.NullInt64
	TotalDiscs  sql.NullInt64
	IndexedAt   sql.NullTime
}

func (da DatabaseAlbum) Hyper(artists []library.HypersonicArtist) *library.HypersonicAlbum {
	var tt *int = nil
	if da.TotalTracks.Valid {
		t := int(da.TotalTracks.Int64)
		tt = &t
	}
	var td *int = nil
	if da.TotalDiscs.Valid {
		t := int(da.TotalDiscs.Int64)
		td = &t
	}
	return &library.HypersonicAlbum{
		Hash:        da.Hash.String,
		Name:        da.Name.String,
		Artwork:     da.Artwork.Bool,
		TotalTracks: tt,
		TotalDiscs:  td,
		IndexedAt:   da.IndexedAt.Time,
		Artists:     artists,
	}
}

type DatabaseTrack struct {
	Hash        string
	Name        string
	AlbumHash   sql.NullString
	Duration    int
	Year        sql.NullInt64
	Genre       sql.NullString
	TrackNumber int
	DiscNumber  int
	FilePath    string
	FileHash    string
	IndexedAt   time.Time
}

func (dt DatabaseTrack) Hyper(album *library.HypersonicAlbum, artists []library.HypersonicArtist) *library.HypersonicTrack {
	var genre *string = nil
	if dt.Genre.Valid {
		genre = &dt.Genre.String
	}
	var year *int = nil
	if dt.Year.Valid {
		y := int(dt.Year.Int64)
		year = &y
	}
	return &library.HypersonicTrack{
		Hash:        dt.Hash,
		Name:        dt.Name,
		Artists:     artists,
		Album:       album,
		TrackNumber: dt.TrackNumber,
		DiscNumber:  dt.DiscNumber,
		Duration:    dt.Duration,
		Year:        year,
		Genre:       genre,
		FilePath:    dt.FilePath,
		FileHash:    dt.FileHash,
		IndexedAt:   dt.IndexedAt,
	}
}

type TrackHashes struct {
	FileHash  string
	TrackHash string
}

type PaginationConfig struct {
	Column    string // use "random" for random order
	Direction string
	Page      int
	Limit     int
}

type PaginatedResult[Item any] struct {
	Items      []Item `json:"items"`
	TotalItems int    `json:"totalItems"`
	TotalPages int    `json:"totalPages"`
}
