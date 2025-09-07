package services

import (
	"mime/multipart"

	"github.com/owenhochwald/harmonia/internal/storage"
)

type MusicService struct {
	Storage storage.Storage
	Repo    string // TODO: change to a Repo struct
}

func (s *MusicService) HandleUpload(fileHeader *multipart.FileHeader) error {
	return nil
}
