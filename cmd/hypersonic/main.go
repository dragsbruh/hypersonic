package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/router"
	"github.com/dragsbruh/hypersonic/internal/router/api/admin"
	"github.com/sirupsen/logrus"
)

func main() {
	config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := database.Init(ctx); err != nil {
		logrus.Fatalf("error in database init: %v", err)
	}

	adminListener, err := net.Listen(config.AdminAddr.Network, config.AdminAddr.Address)
	if err != nil {
		logrus.Fatalf("error listening on admin addr: %v", err)
	}
	defer func() {
		adminListener.Close()
		if config.AdminAddr.Network == "unix" {
			os.Remove(config.AdminAddr.Address)
		}
	}()

	adminServer := http.Server{Handler: admin.Router()}
	go func() {
		logrus.Infof("starting admin server on %s", config.AdminAddr.Address)
		if err := adminServer.Serve(adminListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error serving admin server: %v", err)
		}
	}()

	server := http.Server{Addr: config.Addr, Handler: router.Router()}
	go func() {
		logrus.Infof("server starting on address %s", config.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("error in listen and serve: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("error shutting down server: %v", err)
	}
	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("error shutting down admin server: %v", err)
	}
	logrus.Infoln("exiting")
}
