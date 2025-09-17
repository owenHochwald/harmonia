package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
)

type songRepoSQL struct {
	DB *sql.DB
}

func NewSongRepo(db *sql.DB) SongRepo {
	return &songRepoSQL{DB: db}
}

func (s songRepoSQL) FindById(id string) (*models.Song, error) {
	query := `
		SELECT s.id, s.title, s.artist, s.album, s.year, s.s3_key, s.fingerprint, s.created_at
		FROM songs s
		WHERE s.id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var song models.Song

	if err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&song.ID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Year,
		&song.S3Key,
		&song.Fingerprint,
		&song.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Database error:", err)
		return nil, err
	}
	return &song, nil
}

func (s songRepoSQL) SaveSong(song models.Song) error {
	// Validate required fields
	if err := validateSong(song); err != nil {
		return err
	}

	query := `
		INSERT INTO songs (id, title, artist, album, year, s3_key, fingerprint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.DB.ExecContext(ctx, query,
		song.ID,
		song.Title,
		song.Artist,
		song.Album,
		song.Year,
		song.S3Key,
		song.Fingerprint,
		song.CreatedAt,
	)

	if err != nil {
		fmt.Println("Database error:", err)
		return err
	}

	return nil
}

func validateSong(song models.Song) error {
	if strings.TrimSpace(song.ID) == "" {
		return errors.New("song ID is required")
	}

	if strings.TrimSpace(song.Title) == "" {
		return errors.New("song title is required")
	}

	if strings.TrimSpace(song.Artist) == "" {
		return errors.New("song artist is required")
	}

	if strings.TrimSpace(song.S3Key) == "" {
		return errors.New("S3 key is required")
	}

	if song.Year < 1800 || song.Year > time.Now().Year()+1 {
		return errors.New("invalid year: must be between 1800 and current year")
	}

	if song.CreatedAt.IsZero() {
		return errors.New("created_at timestamp is required")
	}

	return nil
}

func (s songRepoSQL) FindByFingerprint(hash string) (*models.Song, error) {
	query := `
		SELECT s.id, s.title, s.artist, s.album, s.year, s.s3_key, s.fingerprint, s.created_at
		FROM songs s
		JOIN fingerprints f ON s.id = f.song_id::text
		WHERE f.hash = $1
		LIMIT 1
		`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var song models.Song

	if err := s.DB.QueryRowContext(ctx, query, hash).Scan(
		&song.ID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Year,
		&song.S3Key,
		&song.Fingerprint,
		&song.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Database error:", err)
		return nil, err
	}

	return &song, nil
}
