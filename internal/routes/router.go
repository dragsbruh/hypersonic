package routes

import (
	"net/http"
	"sync"

	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/media"
	"github.com/dragsbruh/hypersonic/internal/routes/middleware"
)

func Router(db *database.Database) http.Handler {
	handler := Handler{
		DB: db,
	}
	mediaHandler := media.Handler{
		DB:    db,
		Cache: sync.Map{},
	}

	r := http.NewServeMux()

	r.HandleFunc("GET /library/tracks", handler.GetPaginatedTracks)
	r.HandleFunc("GET /library/tracks/{hash}", handler.GetTrack)

	r.HandleFunc("GET /library/albums", handler.GetPaginatedAlbums)
	r.HandleFunc("GET /library/albums/{hash}", handler.GetAlbum)
	r.HandleFunc("GET /library/albums/{hash}/tracks", handler.GetAlbumTracks)

	r.HandleFunc("GET /library/artists", handler.GetPaginatedArtists)
	r.HandleFunc("GET /library/artists/{hash}", handler.GetArtist)
	r.HandleFunc("GET /library/artists/{hash}/tracks", handler.GetArtistTracks)
	r.HandleFunc("GET /library/artists/{hash}/albums", handler.GetArtistAlbums)

	r.HandleFunc("GET /auth", handler.Me)

	r.HandleFunc("GET /media/tracks/{hash}", mediaHandler.ServeTrack)

	root := http.NewServeMux()
	root.Handle("/", middleware.MustAuth(r))

	root.HandleFunc("POST /auth/login", handler.Login)
	root.HandleFunc("POST /auth/register", handler.Register)

	root.HandleFunc("GET /server/conf", handler.Conf)

	root.HandleFunc("GET /media/albums/{hash}/{res}", mediaHandler.ServeArtwork)

	return http.StripPrefix("/api", root)
}
