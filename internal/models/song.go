package models

import "time"

type Song struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Artist      string    `json:"artist" db:"artist"`
	Album       string    `json:"album" db:"album"`
	Year        int       `json:"year" db:"year"`
	S3Key       string    `json:"s3_key" db:"s3_key"`
	Fingerprint []byte    `json:"fingerprint" db:"fingerprint"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
