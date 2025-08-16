-- name: CreateArtist :exec
INSERT INTO artist (hash, name)
VALUES ($1, $2)
ON CONFLICT (hash) DO UPDATE SET
    name = EXCLUDED.name;

-- name: DeleteArtist :exec
DELETE FROM artist
WHERE artist.hash = $1;

-- name: GetTrackArtists :many
SELECT artist.* FROM track_artist
JOIN artist ON artist.hash = track_artist.artist_hash
WHERE track_artist.track_hash = $1;

-- name: GetAlbumArtists :many
SELECT artist.* FROM album_artist
JOIN artist ON artist.hash = album_artist.artist_hash
WHERE album_artist.album_hash = $1;

-- name: MarkAlbumArtist :exec
INSERT INTO album_artist (album_hash, artist_hash)
VALUES ($1, $2)
ON CONFLICT (album_hash, artist_hash) DO NOTHING;

-- name: MarkTrackArtist :exec
INSERT INTO track_artist (track_hash, artist_hash)
VALUES ($1, $2)
ON CONFLICT (track_hash, artist_hash) DO NOTHING;
