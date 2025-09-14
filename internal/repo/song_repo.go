package repo

import (
	"database/sql"

	"github.com/owenhochwald/harmonia/internal/models"
)

type songRepoSQL struct {
	DB *sql.DB
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
	//TODO implement me
	panic("implement me")
}

func newSongRepo(db *sql.DB) *songRepoSQL {
	return &songRepoSQL{DB: db}
}
