package admin

import (
	"net/http"
	"sync"

	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/library/scanner"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Router(db *database.Database) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /admin/scan", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "upgrade error", http.StatusInternalServerError)
			return
		}
		defer c.Close()

		m := sync.Mutex{}
		err = scanner.ScanAndUpdate(r.Context(), db, func(ss scanner.ScanStatus) {
			m.Lock()
			defer m.Unlock()

			c.WriteJSON(ss)
		})
		if err != nil {
			logrus.Errorf("error while scanning: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})

	return r
}
