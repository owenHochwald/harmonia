package repo

import "github.com/owenhochwald/harmonia/internal/models"

type SongRepo interface {
	SaveSong(song models.Song) error
	SaveFingerprint(fp models.Fingerprint) error
	FindByFingerprint(hash string) ([]models.Song, error)
}
