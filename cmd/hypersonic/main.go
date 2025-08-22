package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/routes"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logrus.Debugf("loading config")
	if err := config.LoadConfig(); err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	logrus.Debugf("initializing database")
	db, err := database.Init(ctx)
	if err != nil {
		logrus.Fatalf("error initializing database: %v", err)
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: routes.Router(db),
	}

	go func() {
		logrus.Infof("listening on addr %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatal(err)
		}
	}()

	<-ctx.Done()

	logrus.Infof("shutting down server")
	sctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := server.Shutdown(sctx); err != nil {
		logrus.Fatalf("failed to shutdown server: %v", err)
	}
	logrus.Infof("server closed")
}
