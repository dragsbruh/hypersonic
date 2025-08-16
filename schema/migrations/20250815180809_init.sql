-- +goose Up
CREATE TABLE IF NOT EXISTS app_user (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_login TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS artist(
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS album(
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,

    total_tracks INT NOT NULL,
    total_discs INT NOT NULL
);

CREATE TABLE IF NOT EXISTS track(
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    
    album_hash TEXT REFERENCES album(hash) ON DELETE CASCADE,
    disc_number INT NOT NULL,
    track_number INT NOT NULL,

    year INT NOT NULL,
    genre TEXT NOT NULL,

    file_path TEXT NOT NULL,
    file_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS track_artist(
    track_hash TEXT NOT NULL REFERENCES track(hash) ON DELETE CASCADE,
    artist_hash TEXT NOT NULL REFERENCES artist(hash) ON DELETE CASCADE,
    PRIMARY KEY(track_hash, artist_hash)
);

CREATE TABLE IF NOT EXISTS album_artist(
    album_hash TEXT NOT NULL REFERENCES album(hash) ON DELETE CASCADE,
    artist_hash TEXT NOT NULL REFERENCES artist(hash) ON DELETE CASCADE,
    PRIMARY KEY(album_hash, artist_hash)
);

CREATE INDEX IF NOT EXISTS idx_track_album_hash ON track(album_hash);
CREATE INDEX IF NOT EXISTS idx_album_artist_artist_hash ON album_artist(artist_hash);

-- +goose Down
DROP TABLE IF EXISTS app_user;

DROP TABLE IF EXISTS track_artist;
DROP TABLE IF EXISTS album_artist;
DROP TABLE IF EXISTS track;
DROP TABLE IF EXISTS album;
DROP TABLE IF EXISTS artist;

DROP INDEX IF EXISTS idx_track_album_hash;
DROP INDEX IF EXISTS idx_album_artist_artist_hash;
