package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dragsbruh/hypersonic/internal/admin"
	"github.com/dragsbruh/hypersonic/internal/api"
	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config.Load()

	queries, err := database.Setup(ctx)
	if err != nil {
		logrus.Fatalf("setup database: %v", err)
	}

	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go admin.Server(&wg, ctx, queries)

	wg.Add(1)
	go api.Server(&wg, ctx, queries)
}
