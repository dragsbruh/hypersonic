package database

import (
	"time"

	"github.com/dragsbruh/hypersonic/internal/library"
)

type DatabaseUser struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	LoginAt   *time.Time `json:"loginAt"`
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
	Hash        string
	Name        string
	Artwork     bool
	TotalTracks *int
	TotalDiscs  *int
	IndexedAt   time.Time
}

func (da DatabaseAlbum) Hyper(artists []library.HypersonicArtist) *library.HypersonicAlbum {
	return &library.HypersonicAlbum{
		Hash:        da.Hash,
		Name:        da.Name,
		Artwork:     da.Artwork,
		TotalTracks: da.TotalTracks,
		TotalDiscs:  da.TotalDiscs,
		IndexedAt:   da.IndexedAt,
		Artists:     artists,
	}
}

type DatabaseTrack struct {
	Hash        string
	Name        string
	AlbumHash   *string
	Duration    int
	Year        *int
	Genre       *string
	TrackNumber int
	DiscNumber  int
	FilePath    string
	FileHash    string
	IndexedAt   time.Time
}

func (dt DatabaseTrack) Hyper(album *library.HypersonicAlbum, artists []library.HypersonicArtist) *library.HypersonicTrack {
	return &library.HypersonicTrack{
		Hash:        dt.Hash,
		Name:        dt.Name,
		Artists:     artists,
		Album:       album,
		TrackNumber: dt.TrackNumber,
		DiscNumber:  dt.DiscNumber,
		Duration:    dt.Duration,
		Year:        dt.Year,
		Genre:       dt.Genre,
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
