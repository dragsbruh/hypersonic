package media

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/library/scanner"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Cache sync.Map // <type>_trackhash -> path/null
	DB    *database.Database
}

func (h *Handler) ServeArtwork(w http.ResponseWriter, r *http.Request) {
	albumHash := r.PathValue("hash")
	cacheKey := fmt.Sprintf("album_%s", albumHash)
	resolutionStr := r.PathValue("res")

	if resolutionStr == "" {
		resolutionStr = "512"
	}
	resolution, err := strconv.Atoi(resolutionStr)
	if err != nil {
		http.Error(w, "bad resolution", http.StatusBadRequest)
		return
	}
	if !slices.Contains(scanner.ArtworkResolutions, resolution) {
		http.Error(w, "unsupported resolution", http.StatusBadRequest)
		return
	}

	basePath, ok := h.Cache.Load(cacheKey)
	if !ok {
		album, err := h.DB.GetAlbum(r.Context(), albumHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "album not found", http.StatusNotFound)
				h.Cache.Store(cacheKey, nil)
			} else {
				logrus.Errorf("error serving artwork: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		if album.Artwork {
			p := filepath.Join(config.DataDir, "albums", albumHash)
			basePath = &p
		} else {
			basePath = nil
		}
		h.Cache.Store(cacheKey, basePath)
	}

	if basePath == nil {
		http.Error(w, "album not found or missing artwork", http.StatusNotFound)
		return
	}

	artworkPath := filepath.Join(*basePath.(*string), fmt.Sprintf("%d.webp", resolution))
	http.ServeFile(w, r, artworkPath)
}

func (h *Handler) ServeTrack(w http.ResponseWriter, r *http.Request) {
	trackHash := r.PathValue("hash")
	cacheKey := fmt.Sprintf("track_%s", trackHash)

	trackPath, ok := h.Cache.Load(cacheKey)
	if !ok {
		track, err := h.DB.GetTrack(r.Context(), trackHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "track not found", http.StatusNotFound)
			} else {
				logrus.Errorf("error serving track: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		trackPath = &track.FilePath
		h.Cache.Store(cacheKey, trackPath)
	}

	http.ServeFile(w, r, *trackPath.(*string))
}
