package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dragsbruh/hypersonic/internal/auth"
	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	DB *database.Database
}

func (h *Handler) GetPaginatedTracks(w http.ResponseWriter, r *http.Request) {
	cfg, err := ParsePagination(r, map[string]bool{
		"name":       true,
		"duration":   true,
		"year":       true,
		"genre":      true,
		"random":     true,
		"indexed_at": true,
	}, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := h.DB.GetPaginatedTracks(r.Context(), *cfg)
	if err != nil {
		logrus.Errorf("error getting paginated tracks: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	successJsonify(w, results)
}

func (h *Handler) GetTrack(w http.ResponseWriter, r *http.Request) {
	trackHash := r.PathValue("hash")
	track, err := h.DB.GetTrack(r.Context(), trackHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			logrus.Errorf("error getting track: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, track)
}

func (h *Handler) GetPaginatedAlbums(w http.ResponseWriter, r *http.Request) {
	cfg, err := ParsePagination(r, map[string]bool{
		"name":         true,
		"total_tracks": true,
		"total_discs":  true,
		"indexed_at":   true,
	}, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := h.DB.GetPaginatedAlbums(r.Context(), *cfg)
	if err != nil {
		logrus.Errorf("error getting paginated albums: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	successJsonify(w, results)
}

func (h *Handler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	albumHash := r.PathValue("hash")
	album, err := h.DB.GetAlbum(r.Context(), albumHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			logrus.Errorf("error getting album: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, album)
}

func (h *Handler) GetAlbumTracks(w http.ResponseWriter, r *http.Request) {
	albumHash := r.PathValue("hash")
	tracks, err := h.DB.GetTracksByAlbum(r.Context(), albumHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "item not found", http.StatusNotFound)
		} else {
			logrus.Errorf("error getting album tracks: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, tracks)
}

func (h *Handler) GetPaginatedArtists(w http.ResponseWriter, r *http.Request) {
	cfg, err := ParsePagination(r, map[string]bool{
		"name": true,
	}, "name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := h.DB.GetPaginatedArtists(r.Context(), *cfg)
	if err != nil {
		logrus.Errorf("error getting paginated artists: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	successJsonify(w, results)
}

func (h *Handler) GetArtist(w http.ResponseWriter, r *http.Request) {
	artistHash := r.PathValue("hash")

	artist, err := h.DB.GetArtist(r.Context(), artistHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "item not found", http.StatusNotFound)
		} else {
			logrus.Errorf("error getting artist: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, artist)
}

func (h *Handler) GetArtistTracks(w http.ResponseWriter, r *http.Request) {
	artistHash := r.PathValue("hash")

	tracks, err := h.DB.GetTracksByArtist(r.Context(), artistHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "item not found", http.StatusNotFound)
		} else {
			logrus.Errorf("error getting artist tracks: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, tracks)
}

func (h *Handler) GetArtistAlbums(w http.ResponseWriter, r *http.Request) {
	artistHash := r.PathValue("hash")

	albums, err := h.DB.GetAlbumsByArtist(r.Context(), artistHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "item not found", http.StatusNotFound)
		} else {
			logrus.Errorf("error getting artist alums: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, albums)
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (l *LoginBody) Validate() error {
	l.Username = strings.TrimSpace(l.Username)
	l.Password = strings.TrimSpace(l.Password)

	// TODO: better validation
	if l.Username == "" || l.Password == "" {
		return fmt.Errorf("username or password is empty")
	}

	return nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()

	body := LoginBody{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	user, err := h.DB.LoginUser(r.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
		} else {
			logrus.Errorf("error while login user: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := h.DB.UpdateUserLogin(ctx, user.ID); err != nil {
		logrus.Errorf("update login: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := auth.CreateJWTUser(user.ID, time.Hour)
	if err != nil {
		logrus.Errorf("error creating jwt user: %v", err)
		http.Error(w, "jwt error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(200)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !config.SelfRegistration {
		http.Error(w, "self registration closed", http.StatusServiceUnavailable)
		return
	}

	defer r.Body.Close()

	body := LoginBody{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	exists, err := h.DB.UserExists(ctx, body.Username)
	if exists {
		http.Error(w, "user with username already exists", http.StatusConflict)
		return
	}

	user, err := h.DB.CreateUser(ctx, body.Username, body.Password)
	if err != nil {
		logrus.Errorf("error while register user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	successJsonify(w, user)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(auth.AuthUserID).(string)

	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.SetCookie(w, &http.Cookie{
				Name:     auth.CookieName,
				HttpOnly: true,
				Expires:  time.Now().Add(-time.Second),
				MaxAge:   -1,
			})
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		} else {
			logrus.Errorf("error while getting user (me): %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	successJsonify(w, user)
}

func (h *Handler) Conf(w http.ResponseWriter, r *http.Request) {
	successJsonify(w, map[string]any{
		"register": config.SelfRegistration,
	})
}

func successJsonify(w http.ResponseWriter, o any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(o)
}
