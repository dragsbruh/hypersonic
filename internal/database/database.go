package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/dragsbruh/hypersonic/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// stolen cutely from sqlc
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Database struct {
	Pool *pgxpool.Pool
	tx   pgx.Tx
}

func (db *Database) Begin(ctx context.Context) (*Database, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	return &Database{
		Pool: db.Pool,
		tx:   tx,
	}, nil
}

func (db *Database) DBTX() DBTX {
	if db.tx == nil {
		return db.Pool
	}
	return db.tx
}

func (db *Database) Rollback(ctx context.Context) error {
	if db.tx == nil {
		return fmt.Errorf("not a tx")
	}
	if err := db.tx.Rollback(ctx); err != nil {
		return err
	}
	return nil
}

func (db *Database) Commit(ctx context.Context) error {
	if db.tx == nil {
		return fmt.Errorf("not a tx")
	}

	if err := db.tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (db *Database) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.DBTX().Exec(ctx, sql, args...)
}

func (db *Database) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.DBTX().Query(ctx, sql, args...)
}

func (db *Database) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.DBTX().QueryRow(ctx, sql, args...)
}

func Migrate(ctx context.Context) error {
	db, err := sql.Open("pgx", config.DatabaseUrl)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
	return nil
}

func Init(ctx context.Context) (*Database, error) {
	if err := Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	pool, err := pgxpool.New(ctx, config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	db := &Database{
		Pool: pool,
		tx:   nil,
	}

	return db, nil
}

//go:embed migrations/*.sql
var embedMigrations embed.FS
