package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
)

type songRepoSQL struct {
	DB *sql.DB
}

func newSongRepo(db *sql.DB) *songRepoSQL {
	return &songRepoSQL{DB: db}
}

func (s songRepoSQL) SaveSong(song models.Song) error {
	//TODO implement me
	panic("implement me")
}

func (s songRepoSQL) SaveFingerprint(fp models.Fingerprint) error {
	//TODO implement me
	panic("implement me")
}

func (s songRepoSQL) FindByFingerprint(hash string) ([]models.Song, error) {
	query := `
			SELECT s.id, s.title, s.artist, s.album, s.year, s.fingerprint
			FROM songs s
			JOIN fingerprints f ON s.id = f.song_id
			WHERE f.hash = $1
			`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query, hash)
	if err != nil {
		fmt.Println("Database error:", err)
		return nil, err
	}
	defer rows.Close()

	songs := []models.Song{}

	for rows.Next() {
		var song models.Song
		err := rows.Scan(
			&song.ID,
			&song.Title,
			&song.Artist,
			&song.Album,
			&song.Year,
			&song.Fingerprint,
		)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(songs) == 0 {
		fmt.Println("No found records")
	}
	return songs, nil

}
