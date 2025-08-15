package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dragsbruh/hypersonic/internal/database"
	"github.com/dragsbruh/hypersonic/internal/database/queries"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid"
)

type User struct {
	ID           string
	Username     string
	CreatedAt    time.Time
	LastLogin    time.Time
	PasswordEnc  []byte
	PasswordHash string
}

func CreateUser(ctx context.Context, username, password string) (string, error) {
	now := time.Now().UTC()
	userID := ulid.MustNew(ulid.Timestamp(now), rand.Reader).String()

	hash, err := PasswordHashFunction(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	enc, err := PasswordEncryptFunction(password)
	if err != nil {
		return "", fmt.Errorf("encrypt password: %w", err)
	}

	err = database.Queries.CreateUser(ctx, queries.CreateUserParams{
		ID:           userID,
		Username:     username,
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		PasswordHash: hash,
		PasswordEnc:  enc,
	})
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	return userID, nil
}

func GetUserByName(ctx context.Context, username string) (*User, error) {
	user, err := database.Queries.GetUserByName(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by name: %w", err)
	}

	return &User{
		ID:           user.ID,
		Username:     user.Username,
		CreatedAt:    user.CreatedAt.Time, // probably local it? idk also in getuserbyid
		LastLogin:    user.LastLogin.Time,
		PasswordHash: user.PasswordHash,
		PasswordEnc:  user.PasswordEnc,
	}, nil
}

func GetUserByID(ctx context.Context, id string) (*User, error) {
	user, err := database.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by name: %w", err)
	}

	return &User{
		ID:           user.ID,
		Username:     user.Username,
		CreatedAt:    user.CreatedAt.Time, // same here as above
		LastLogin:    user.LastLogin.Time,
		PasswordHash: user.PasswordHash,
		PasswordEnc:  user.PasswordEnc,
	}, nil
}

func UpdateUserEncPassword(ctx context.Context, id, password string) error {
	enc, err := PasswordEncryptFunction(password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}

	err = database.Queries.UpdateUserEncPassword(ctx, queries.UpdateUserEncPasswordParams{
		ID:          id,
		PasswordEnc: enc,
	})
	if err != nil {
		return fmt.Errorf("update password db: %w", err)
	}

	return nil
}

func CheckUserLogin(ctx context.Context, username, password string) (*User, error) {
	hash, err := PasswordHashFunction(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := GetUserByName(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by name: %w", err)
	}

	if user.PasswordHash != hash {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
