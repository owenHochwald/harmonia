package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
)

type fingerprintRepoSQL struct {
	DB *sql.DB
}

func NewFingerprintRepo(db *sql.DB) FingerprintRepo {
	return &fingerprintRepoSQL{DB: db}
}

func (f *fingerprintRepoSQL) SaveFingerprint(fingerprint models.Fingerprint) error {
	query := `
		INSERT INTO fingerprints (song_id, hash, time_offset)
		VALUES ($1, $2, $3)
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := f.DB.ExecContext(ctx, query,
		fingerprint.SongID,
		fingerprint.Hash,
		fingerprint.TimeOffset,
	)

	if err != nil {
		fmt.Println("Database error:", err)
		return err
	}

	return nil
}

func (f *fingerprintRepoSQL) FindByHash(hash string) (*models.Fingerprint, error) {
	query := `
		SELECT id, song_id, hash, time_offset
		FROM fingerprints
		WHERE hash = $1
		LIMIT 1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var fingerprint models.Fingerprint

	if err := f.DB.QueryRowContext(ctx, query, hash).Scan(
		&fingerprint.ID,
		&fingerprint.SongID,
		&fingerprint.Hash,
		&fingerprint.TimeOffset,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Database error:", err)
		return nil, err
	}

	return &fingerprint, nil
}

func (f *fingerprintRepoSQL) FindById(id int64) (*models.Fingerprint, error) {
	query := `
		SELECT id, song_id, hash, time_offset
		FROM fingerprints
		WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var fingerprint models.Fingerprint

	if err := f.DB.QueryRowContext(ctx, query, id).Scan(
		&fingerprint.ID,
		&fingerprint.SongID,
		&fingerprint.Hash,
		&fingerprint.TimeOffset,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Database error:", err)
		return nil, err
	}

	return &fingerprint, nil
}

func (f *fingerprintRepoSQL) FindBySongId(songId string) (*models.Fingerprint, error) {
	query := `
		SELECT id, song_id, hash, time_offset
		FROM fingerprints
		WHERE song_id = $1
		LIMIT 1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var fingerprint models.Fingerprint

	if err := f.DB.QueryRowContext(ctx, query, songId).Scan(
		&fingerprint.ID,
		&fingerprint.SongID,
		&fingerprint.Hash,
		&fingerprint.TimeOffset,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println("Database error:", err)
		return nil, err
	}

	return &fingerprint, nil
}
