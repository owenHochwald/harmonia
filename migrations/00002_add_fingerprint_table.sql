-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

DROP TABLE IF EXISTS song CASCADE;

CREATE TABLE songs
(
    id          SERIAL PRIMARY KEY,
    title       TEXT      NOT NULL,
    artist      TEXT      NOT NULL,
    album       TEXT      NOT NULL,
    year        INTEGER   NOT NULL,
    s3_key      TEXT      NOT NULL,
    fingerprint BYTEA     NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fingerprints
(
    id      SERIAL PRIMARY KEY,
    song_id INTEGER NOT NULL,
    hash    INTEGER NOT NULL,
    time_offset  INTEGER NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs (id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS fingerprints;
-- +goose StatementEnd
