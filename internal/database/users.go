package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
)

func (db *Database) GetUserByID(ctx context.Context, id string) (*DatabaseUser, error) {
	row := db.QueryRow(ctx, `
		SELECT id, username, created_at, login_at
		FROM app_user
		WHERE id = $1
	`, id)
	user := DatabaseUser{}
	if err := row.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LoginAt); err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}

	return &user, nil
}

func (db *Database) CreateUser(ctx context.Context, username, password string) (*DatabaseUser, error) {
	now := time.Now()
	user := DatabaseUser{
		ID:        ulid.MustNew(ulid.Timestamp(now), rand.Reader).String(),
		Username:  username,
		Password:  password,
		CreatedAt: now,
		LoginAt:   sql.NullTime{Valid: false},
	}

	_, err := db.Exec(ctx, `
		INSERT INTO app_user (id, username, password, created_at, login_at)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Username, user.Password, user.CreatedAt, user.LoginAt)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}

	user.Password = ""
	return &user, nil
}

func (db *Database) UpdateUserLogin(ctx context.Context, userID string) error {
	_, err := db.Exec(ctx, `
		UPDATE app_user
		SET login_at = now()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (db *Database) UpdateUserPassword(ctx context.Context, userID, password string) error {
	_, err := db.Exec(ctx, `
		UPDATE app_user
		SET password = $2
		WHERE id = $1
	`, userID, password)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (db *Database) LoginUser(ctx context.Context, username, password string) (*DatabaseUser, error) {
	row := db.QueryRow(ctx, `
		SELECT id, username, created_at, login_at
		FROM app_user
		WHERE username = $1 AND password = $2
	`, username, password)
	user := DatabaseUser{}
	if err := row.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.LoginAt); err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}

	return &user, nil
}

func (db *Database) UserExists(ctx context.Context, username string) (bool, error) {
	row := db.QueryRow(ctx, `SELECT id FROM app_user WHERE username = $1`, username)

	userID := ""
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query row: %w", err)
	}

	return true, nil
}
