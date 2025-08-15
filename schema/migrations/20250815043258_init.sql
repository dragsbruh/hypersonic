-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_user (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_login TIMESTAMPTZ NOT NULL,
    password_enc BYTEA NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artist (
    id TEXT PRIMARY KEY, -- artist hash
    title TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS album (
    id TEXT PRIMARY KEY, -- album hash
    title TEXT NOT NULL,
    total_discs INT NOT NULL,
    total_tracks INT NOT NULL
);

CREATE TABLE IF NOT EXISTS track (
    id TEXT PRIMARY KEY, -- track hash
    title TEXT NOT NULL,
    album_id TEXT REFERENCES album(id) ON DELETE CASCADE,
    disc_number INT NOT NULL,
    track_number INT NOT NULL,
    year INT NOT NULL,
    genre TEXT NOT NULL,
    composer TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS album_artist (
    album_id TEXT NOT NULL REFERENCES album(id) ON DELETE CASCADE,
    artist_id TEXT NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    PRIMARY KEY (album_id, artist_id)
);

CREATE TABLE IF NOT EXISTS track_artist (
    track_id TEXT NOT NULL REFERENCES track(id) ON DELETE CASCADE,
    artist_id TEXT NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, artist_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_user;

DROP TABLE IF EXISTS track_artist;
DROP TABLE IF EXISTS album_artist;
DROP TABLE IF EXISTS track;
DROP TABLE IF EXISTS album;
DROP TABLE IF EXISTS artist;
-- +goose StatementEnd
