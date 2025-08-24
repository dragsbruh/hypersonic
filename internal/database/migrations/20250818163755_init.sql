-- +goose Up
CREATE TABLE IF NOT EXISTS app_user (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    login_at TIMESTAMPTZ -- last login
);

CREATE TABLE IF NOT EXISTS artist (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS album (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,

    artwork BOOLEAN NOT NULL,

    total_tracks INT,
    total_discs INT,

    indexed_at TIMESTAMPTZ NOT NULL -- album indexed at
);

CREATE TABLE IF NOT EXISTS track (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    
    album_hash TEXT REFERENCES album(hash) ON DELETE CASCADE,
    track_number INT NOT NULL,
    disc_number INT NOT NULL,
    
    duration INT NOT NULL,
    year INT,
    genre TEXT,

    file_path TEXT NOT NULL,
    file_hash TEXT NOT NULL,

    indexed_at TIMESTAMPTZ NOT NULL
);


CREATE TABLE IF NOT EXISTS track_artist (
    track_hash TEXT NOT NULL REFERENCES track(hash) ON DELETE CASCADE,
    artist_hash TEXT NOT NULL REFERENCES artist(hash) ON DELETE CASCADE,
    PRIMARY KEY (track_hash, artist_hash)
);

CREATE TABLE IF NOT EXISTS album_artist (
    album_hash TEXT NOT NULL REFERENCES album(hash) ON DELETE CASCADE,
    artist_hash TEXT NOT NULL REFERENCES artist(hash) ON DELETE CASCADE,
    PRIMARY KEY (album_hash, artist_hash)
);

CREATE INDEX IF NOT EXISTS idx_track_album_hash ON track(album_hash);

-- +goose Down
DROP TABLE IF EXISTS track_artist;
DROP TABLE IF EXISTS album_artist;
DROP TABLE IF EXISTS track;
DROP TABLE IF EXISTS album;
DROP TABLE IF EXISTS artist;

DROP TABLE IF EXISTS app_user;