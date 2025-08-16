package admin

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/sirupsen/logrus"
)

func Server(wg *sync.WaitGroup, ctx context.Context, queries *database.PGQueries) {
	defer wg.Done()

	listener, err := net.Listen(config.ADMIN_NETWORK, config.ADMIN_ADDR)
	if err != nil {
		logrus.Errorf("error listening admin: %v", err)
		return
	}

	defer func() {
		listener.Close()
		if config.ADMIN_NETWORK == "unix" {
			os.Remove(config.ADMIN_ADDR)
		}
	}()

	server := &http.Server{
		Handler: Router(ctx, queries),
	}

	errCh := make(chan error, 1)
	go func() {
		logrus.Infof("starting admin server on addr %s over %s", config.ADMIN_ADDR, config.ADMIN_NETWORK)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("error shutting down admin: %v", err)
		}
		logrus.Infof("admin server shut down")

	case err := <-errCh:
		logrus.Errorf("error serving: %v", err)
	}
}
