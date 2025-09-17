package services

import (
	"context"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
)

type MockStorage struct{}

func (m MockStorage) Upload(ctx context.Context, key string, data []byte) error {
	return nil
}

func (m MockStorage) Download(ctx context.Context, key string) ([]byte, error) {
	return []byte("mock data"), nil
}

func (m MockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

type MockRepository struct{}

func (m MockRepository) FindById(id int) error {
	//TODO implement me
	panic("implement me")
}

func (m MockRepository) SaveSong(song models.Song) error {
	return nil
}

func (m MockRepository) SaveFingerprint(fp models.Fingerprint) error {
	return nil
}

func (m MockRepository) FindByFingerprint(hash string) ([]models.Song, error) {
	return []models.Song{MockSongFactory()}, nil
}

func MockSongFactory() models.Song {
	return models.Song{
		ID:          "1",
		Title:       "title",
		Artist:      "artist",
		Album:       "album",
		Year:        2025,
		S3Key:       "s3key",
		Fingerprint: []byte("fingerprint"),
		CreatedAt:   time.Now(),
	}
}
