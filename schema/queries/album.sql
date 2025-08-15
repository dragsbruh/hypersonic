-- name: CreateAlbum :exec
INSERT INTO album (id, title, total_discs, total_tracks)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    total_discs = EXCLUDED.total_discs,
    total_tracks = EXCLUDED.total_tracks;

-- name: GetAlbumArtists :many
SELECT * FROM album_artist
JOIN artist ON artist.id = album_artist.artist_id
WHERE album_artist.album_id = $1;

-- name: GetAlbumByHash :one
SELECT * FROM album
WHERE album.id = $1
LIMIT 1;

-- name: GetAlbums :many
SELECT * FROM album;

-- name: MarkAlbumArtist :exec
INSERT INTO album_artist (album_id, artist_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;
