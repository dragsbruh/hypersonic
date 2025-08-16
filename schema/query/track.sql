-- name: CreateTrack :exec
INSERT INTO track (
    hash,
    name,
    album_hash,
    disc_number,
    track_number,
    year,
    genre,
    file_path,
    file_hash
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (hash) DO UPDATE SET
    name = EXCLUDED.name,
    album_hash = EXCLUDED.album_hash,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    year = EXCLUDED.year,
    genre = EXCLUDED.genre,
    file_path = EXCLUDED.file_path,
    file_hash = EXCLUDED.file_hash;

-- name: GetTrackByHash :one
SELECT * FROM track
WHERE track.hash = $1
LIMIT 1;

-- name: GetPaginatedTracks :many
SELECT * FROM track
LIMIT $1 OFFSET $2;

-- name: GetAlbumTracks :many
SELECT * FROM track
WHERE track.album_hash = $1
ORDER BY disc_number, track_number;

-- name: GetArtistTracks :many
SELECT track.* FROM track_artist
JOIN track ON track.hash = track_artist.track_hash
WHERE track_artist.artist_hash = $1;

-- name: GetTrackHashes :many
SELECT hash, file_hash FROM track;

-- name: DeleteTrack :exec
DELETE FROM track WHERE track.hash = $1;
