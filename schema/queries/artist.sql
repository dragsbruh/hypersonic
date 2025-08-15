-- name: CreateArtist :exec
INSERT INTO artist (id, title)
VALUES ($1, $2) ON CONFLICT (id) DO NOTHING;

-- name: GetArtistByHash :one
SELECT * FROM artist
WHERE artist.id = $1
LIMIT 1;

-- name: GetArtists :many
SELECT * FROM artist;
