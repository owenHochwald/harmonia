package services

import (
	"context"
	"testing"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/owenhochwald/harmonia/internal/repo"
	"github.com/stretchr/testify/assert"
)

var (
	ctx      = context.Background()
	service  *MusicService
	testSong models.Song
)

func setupService() (*MusicService, context.Context, models.Song) {
	service = &MusicService{
		Storage: MockStorage{},
		Repo:    &repo.MockSongRepo{},
	}
	testSong = MockSongFactory()

	return service, ctx, testSong
}

func TestHandleUpload_Success(t *testing.T) {
	service, ctx, testSong := setupService()
	err := service.HandleUpload(ctx, testSong, []byte("test data"))
	assert.Nil(t, err)
	assert.NoError(t, err)
}

func TestHandleUpload_Fail_EmptyData(t *testing.T) {
	service, ctx, testSong := setupService()
	err := service.HandleUpload(ctx, testSong, nil)
	assert.ErrorContains(t, err, "Empty data")
}

func TestHandleUpload_Fail_MalformedSong(t *testing.T) {
	service, ctx, _ := setupService()
	err := service.HandleUpload(ctx, models.Song{}, nil)
	assert.ErrorContains(t, err, "Malformed song")
}

func TestMusicService_Identify_Success(t *testing.T) {
	service, ctx, testSong := setupService()

	songs, err := service.Identify(ctx, "test song fingerprint")
	assert.Nil(t, err)
	assert.Equal(t, testSong, songs[0])
}

func TestMusicService_Identify_Fail_MissingRecord(t *testing.T) {
	service, ctx, testSong := setupService()

	songs, err := service.Identify(ctx, "missing song")
	assert.ErrorContains(t, err, "No found record")
	assert.NotEqual(t, testSong, songs[0])
}

func TestMusicService_Identify_Fail_MalformedHash(t *testing.T) {
	service, ctx, testSong := setupService()

	songs, err := service.Identify(ctx, "")
	assert.ErrorContains(t, err, "Malformed hash")
	assert.NotEqual(t, testSong, songs[0])
}
