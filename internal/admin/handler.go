package admin

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/library"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Queries *database.PGQueries
	MainCtx context.Context
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// opens websocket to scanner. transaction aborts if connection closes and scan is reverted
func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(h.MainCtx)
	defer cancel()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("websocket upgrade error: %v", err)
		http.Error(w, "upgrade error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	defer conn.WriteMessage(websocket.CloseNormalClosure, []byte("scan complete"))

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if _, _, err := conn.ReadMessage(); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	existingHashes, err := database.GetExistingHashes(ctx, h.Queries)
	if err != nil {
		logrus.Errorf("error getting existing hashes: %v", err)
		sendWsError(conn, "get existing hashes", err)
		return
	}

	var seenTrackHashes = map[string]struct{}{}

	err = library.ScanDirectory(ctx, config.MUSIC_DIR, func(scanned int, trackPath string) error {
		// TODO: actual scan and write to db transaction
		fileHash, err := library.GetFileHash(trackPath)
		if err != nil {
			return fmt.Errorf("file hash: %w", err)
		}
		trackHash, ok := existingHashes[fileHash]
		if ok {
			conn.WriteJSON(map[string]any{
				"status":    "skip",
				"fileHash":  fileHash,
				"trackHash": trackHash,
				"trackPath": trackPath,
				"scanned":   scanned,
			})
			seenTrackHashes[trackHash] = struct{}{}
			return nil
		}

		track, err := library.ParseTrackMetadata(trackPath, fileHash)
		if err != nil {
			sendWsError(conn, fmt.Sprintf("parse track: %s", trackPath), err)
			return fmt.Errorf("parse track %s: %w", filepath.Base(trackPath), err)
		}

		if err := database.CreateTrack(ctx, h.Queries, track); err != nil {
			sendWsError(conn, fmt.Sprintf("create track: %s", trackPath), err)
			return fmt.Errorf("create track: %w", err)
		}

		conn.WriteJSON(map[string]any{
			"status":    "add",
			"trackHash": track.Hash,
			"count":     scanned,
			"title":     track.Name,
		})

		return nil
	})
	if err != nil {
		logrus.Errorf("scan failed: %v", err)
		return
	}

	conn.WriteJSON(map[string]any{
		"status": "complete",
	})
}

func sendWsError(conn *websocket.Conn, state string, err error) {
	conn.WriteJSON(map[string]string{
		"status": "error",
		"error":  fmt.Sprintf("%s: %s", state, err.Error()),
	})
}
