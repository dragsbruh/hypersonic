-- name: CreateUser :exec
INSERT INTO app_user (id, username, created_at, password_enc, password_hash, last_login)
VALUES ($1, $2, $3, $4, $5, now());

-- name: GetUserByID :one
SELECT * FROM app_user
WHERE app_user.id = $1
LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM app_user
WHERE app_user.username = $1
LIMIT 1;

-- name: UpdateUserEncPassword :exec
UPDATE app_user
SET password_enc = $2,
    last_login = now()
WHERE app_user.id = $1;
