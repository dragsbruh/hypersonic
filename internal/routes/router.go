package routes

import (
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/media"
	"github.com/dragsbruh/hypersonic/internal/routes/admin"
	"github.com/dragsbruh/hypersonic/internal/routes/middleware"
	"github.com/sirupsen/logrus"
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

	root.Handle("/admin/", middleware.MustAdmin(admin.Router(db)))

	actualRoot := http.NewServeMux()

	if config.Production {
		actualRoot.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := filepath.Join("dist", r.URL.Path)

			file, err := staticEmbed.Open(path)
			if os.IsNotExist(err) {
				path = "dist/index.html"
			} else if err != nil {
				logrus.Errorf("error serving file: %v", err)
				return
			} else {
				defer file.Close()
			}

			http.ServeFileFS(w, r, staticEmbed, path)
		})
	}

	actualRoot.Handle("/api/", http.StripPrefix("/api", root))

	return actualRoot
}

//go:embed dist/*
var staticEmbed embed.FS
