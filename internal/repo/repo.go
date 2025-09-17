package repo

import "github.com/owenhochwald/harmonia/internal/models"

type SongRepo interface {
	SaveSong(song models.Song) error
	FindById(id string) (*models.Song, error)
	FindByFingerprint(hash string) (*models.Song, error)
}

type FingerprintRepo interface {
	SaveFingerprint(models.Fingerprint) error
	FindByHash(hash string) (*models.Fingerprint, error)
	FindById(id int64) (*models.Fingerprint, error)
	FindBySongId(songId string) (*models.Fingerprint, error)
}
