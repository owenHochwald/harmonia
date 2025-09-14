-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE song
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS song;
-- +goose StatementEnd