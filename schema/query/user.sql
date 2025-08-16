-- name: CreateUser :exec
INSERT INTO app_user (id, username, password, created_at, last_login)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByID :one
SELECT
    app_user.id,
    app_user.username,
    app_user.created_at,
    app_user.last_login
FROM app_user
WHERE app_user.id = $1
LIMIT 1;

-- name: DeleteUserByID :exec
DELETE FROM app_user
WHERE app_user.id = $1;

-- name: GetUserLoginByHash :one
SELECT
    app_user.id,
    app_user.username,
    app_user.created_at,
    app_user.last_login
FROM app_user
WHERE
    app_user.username = $1 AND md5(app_user.password || $3) = $2
LIMIT 1;

-- for legacy subsonic requests with plaintext passwords (ew)
-- name: GetUserLogin :one
SELECT
    app_user.id,
    app_user.username,
    app_user.created_at,
    app_user.last_login
FROM app_user
WHERE
    app_user.username = $1 AND app_user.password = $2
LIMIT 1;

-- name: MarkUserLogin :exec
UPDATE app_user
SET last_login = now()
WHERE app_user.id = $1;
