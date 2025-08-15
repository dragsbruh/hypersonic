package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/dragsbruh/hypersonic/internal/auth"
	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/media"
	"github.com/dragsbruh/hypersonic/internal/scan"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func Router() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /library/scan", routeScan)
	r.HandleFunc("POST /users/create", routeCreateUser)

	return r
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func routeScan(w http.ResponseWriter, r *http.Request) {
	// TODO: probably take main context here smh?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("failed to upgrade scan websocket: %v", err)
		return
	}
	defer conn.Close()

	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	existingArr, err := media.GetComparableTracks(ctx, database.Queries)
	if err != nil {
		logrus.Errorf("failed to get comparable tracks: %v", err)
		return
	}

	tx, err := database.Pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	queries := database.Queries.WithTx(tx)

	seenHashes := map[string]struct{}{}

	err = scan.RecursiveScanDirectory(ctx, config.MusicDir, func(count int, fullpath, filehash string) error {
		seenHashes[filehash] = struct{}{}

		exists := slices.ContainsFunc(existingArr, func(e media.ComparableTrack) bool {
			return e.FileHash == filehash
		})
		if exists {
			conn.WriteJSON(map[string]any{
				"Status":   "Skip",
				"Count":    count,
				"Path":     fullpath,
				"FileHash": filehash,
			})
			return nil
		}

		track, err := scan.ScanAudioMetadata(fullpath)
		if err != nil {
			return fmt.Errorf("scan track: %w", err)
		}
		track.FileHash = filehash

		if err := media.CreateTrack(ctx, queries, *track); err != nil {
			return fmt.Errorf("create track: %w", err)
		}

		conn.WriteJSON(map[string]any{
			"Status":   "Add",
			"Count":    count,
			"Path":     track.FilePath,
			"FileHash": track.FileHash,
		})
		return nil
	})
	if err != nil {
		logrus.Errorf("scan aborted due to error: %v", err)
		return
	}

	select {
	case <-ctx.Done():
		logrus.Errorf("context cancelled, skipping deletion")
		return
	default:
	}

	for _, track := range existingArr {
		_, seen := seenHashes[track.FileHash]
		if !seen {
			if err := media.DeleteTrack(ctx, queries, track.Hash); err != nil {
				logrus.Errorf("delete track failed: %v", err)
				return
			}
			conn.WriteJSON(map[string]any{
				"Status":   "Delete",
				"Path":     track.FilePath,
				"FileHash": track.FileHash,
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logrus.Errorf("error commiting transaction: %v", err)
		return
	}

	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "scan complete"))
}

type createUserRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

func routeCreateUser(w http.ResponseWriter, r *http.Request) {
	var data createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	data.Username = strings.TrimSpace(data.Username)
	data.Password = strings.TrimSpace(data.Password)

	userID, err := auth.CreateUser(r.Context(), data.Username, data.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"Status": "Error",
			"Error":  err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"Status": "Ok",
		"UserID": userID,
	})
}
