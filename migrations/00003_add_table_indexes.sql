-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE INDEX IF NOT EXISTS idx_fingerprints_hash ON fingerprints(hash);
CREATE INDEX IF NOT EXISTS idx_fingerprints_song_id ON fingerprints(song_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_fingerprints_hash;
DROP INDEX IF EXISTS idx_fingerprints_song_id;
-- +goose StatementEnd
