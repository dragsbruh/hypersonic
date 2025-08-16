package database

import (
	"context"
	"fmt"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/dragsbruh/hypersonic/internal/database/query"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGQueries struct {
	query.Queries
	Pool *pgxpool.Pool
}

func (q *PGQueries) Tx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, *PGQueries, error) {
	tx, err := q.Pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	queries := q.WithTx(tx)
	return tx, &PGQueries{
		Queries: *queries,
		Pool:    q.Pool,
	}, err
}

func Setup(ctx context.Context) (*PGQueries, error) {
	conf, err := pgxpool.ParseConfig(config.DATABASE_URL)
	if err != nil {
		return nil, fmt.Errorf("parse dburl: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return &PGQueries{
		Queries: *query.New(pool),
		Pool:    pool,
	}, nil
}
