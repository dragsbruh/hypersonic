-- name: CreateAlbum :exec
INSERT INTO album (hash, name, total_tracks, total_discs)
VALUES ($1, $2, $3, $4)
ON CONFLICT (hash) DO UPDATE SET
    name = EXCLUDED.name,
    total_tracks = EXCLUDED.total_tracks,
    total_discs = EXCLUDED.total_discs;

-- name: GetAlbumByHash :one
SELECT * FROM album
WHERE album.hash = $1
LIMIT 1;

-- name: GetPaginatedAlbums :many
SELECT * FROM album
LIMIT $1 OFFSET $2;

-- name: GetArtistAlbums :many
SELECT * FROM album_artist
JOIN album ON album.hash = album_artist.album_hash
WHERE album_artist.artist_hash = $1;

-- name: DeleteAlbum :exec
DELETE FROM album WHERE album.hash = $1;
