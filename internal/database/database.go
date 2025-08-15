package database

import (
	"context"
	"fmt"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Queries *queries.Queries
var Pool *pgxpool.Pool

func Init(ctx context.Context) error {
	conf, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("parse dburl: %w", err)
	}

	Pool, err = pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("new pool: %w", err)
	}

	Queries = queries.New(Pool)

	return nil
}
