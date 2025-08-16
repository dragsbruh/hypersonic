package api

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/sirupsen/logrus"
)

func Server(wg *sync.WaitGroup, ctx context.Context, queries *database.PGQueries) {
	defer wg.Done()

	server := http.Server{
		Addr:    config.ADDR,
		Handler: Router(ctx, queries),
	}

	errCh := make(chan error, 1)
	go func() {
		logrus.Infof("starting api server on addr %s", config.ADDR)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("error shutting down api: %v", err)
		}
		logrus.Infof("api server shut down")

	case err := <-errCh:
		logrus.Errorf("error listening api: %v", err)
	}
}
