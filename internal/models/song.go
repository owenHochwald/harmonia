package models

import "time"

type Song struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Artist      string    `db:"artist"`
	Album       string    `db:"album"`
	Year        int       `db:"year"`
	S3Key       string    `db:"s3_key"`
	Fingerprint []byte    `db:"fingerprint"`
	CreatedAt   time.Time `db:"created_at"`
}
