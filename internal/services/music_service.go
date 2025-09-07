package services

import (
	"context"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/owenhochwald/harmonia/internal/repo"
	"github.com/owenhochwald/harmonia/internal/storage"
)

type MusicService struct {
	Storage storage.Storage
	Repo    repo.SongRepo
}

// TODO: Implement full functionality
func (s *MusicService) HandleUpload(ctx context.Context, song models.Song, data []byte) error {
	return nil
}

// TODO: Implement full functionality
func (s *MusicService) Identify(ctx context.Context, hash string) ([]models.Song, error) {
	return nil, nil
}
