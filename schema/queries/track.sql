-- name: CreateTrack :exec
INSERT INTO track
    (
        id,
        title,
        album_id,
        disc_number,
        track_number,
        year,
        genre,
        composer,
        file_path,
        file_hash
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    album_id = EXCLUDED.album_id,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    year = EXCLUDED.year,
    genre = EXCLUDED.genre,
    composer = EXCLUDED.composer,
    file_path = EXCLUDED.file_path,
    file_hash = EXCLUDED.file_hash
;

-- name: GetTrackArtists :many
SELECT * FROM track_artist
JOIN artist ON artist.id = track_artist.artist_id
WHERE track_artist.track_id = $1;

-- name: GetTrackByFilePath :one
SELECT * FROM track
WHERE track.file_path = $1
LIMIT 1;

-- name: GetTrackByHash :one
SELECT * FROM track
WHERE track.id = $1
LIMIT 1;

-- name: GetTracks :many
SELECT * FROM track;

-- name: DeleteTrack :exec
DELETE FROM track
WHERE track.id = $1;

-- name: MarkTrackArtist :exec
INSERT INTO track_artist (track_id, artist_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;
