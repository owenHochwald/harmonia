package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/owenhochwald/harmonia/internal/repo"
	"github.com/owenhochwald/harmonia/internal/storage"
)

type MusicServiceInterface interface {
	HandleUpload(data []byte) (*models.Song, error)
	Identify(ctx context.Context, hash string) ([]models.Song, error)
}

type MusicService struct {
	Storage      storage.Storage
	Repo         repo.SongRepo
	AudioService AudioServiceInterface
}

func NewMusicService(storage storage.Storage, repo repo.SongRepo, audioService AudioServiceInterface) MusicServiceInterface {
	return &MusicService{
		Storage:      storage,
		Repo:         repo,
		AudioService: audioService,
	}
}

func (s *MusicService) HandleUpload(data []byte) (*models.Song, error) {
	_, err := s.AudioService.ReadWAVProperties(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error getting wav file metadata: %w", err)
	}

	processed, err := s.AudioService.ConvertToMono(data)
	if err != nil {
		return nil, fmt.Errorf("error converting wav file: %w", err)
	}
	processed, err = s.AudioService.Resample(processed, 16000)
	if err != nil {
		return nil, fmt.Errorf("error resampling wav file: %w", err)
	}
	processed, err = s.AudioService.Normalize(processed)
	if err != nil {
		return nil, fmt.Errorf("error normalizing wav file: %w", err)
	}

	_, err = s.AudioService.Spectrogram(processed, 2048, 512)
	if err != nil {
		return nil, fmt.Errorf("error getting spectrogram for wav file: %w", err)
	}
	song := models.Song{}

	//song := models.Song{
	//	// add title
	//	// add artist
	//	// add album
	//	// add year
	//
	//}

	s.Repo.SaveSong(song)

	return &song, nil
}

// TODO: Implement full functionality
func (s *MusicService) Identify(ctx context.Context, hash string) ([]models.Song, error) {
	return nil, nil
}
