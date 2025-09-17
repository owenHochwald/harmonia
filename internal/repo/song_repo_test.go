package repo

import (
	"testing"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSongRepo_SaveSong(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewSongRepo(db)

	tests := []struct {
		name    string
		song    models.Song
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid song",
			song: models.Song{
				ID:          "123", // Match fingerprint SongID
				Title:       "Test Song",
				Artist:      "Test Artist",
				Album:       "Test Album",
				Year:        2023,
				S3Key:       "songs/test-song.mp3",
				Fingerprint: []byte("test-fingerprint"),
				CreatedAt:   time.Now().UTC(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			song: models.Song{
				Title:     "Test Song",
				Artist:    "Test Artist",
				S3Key:     "songs/test-song.mp3",
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "song ID is required",
		},
		{
			name: "missing title",
			song: models.Song{
				ID:        "song-123",
				Artist:    "Test Artist",
				S3Key:     "songs/test-song.mp3",
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "song title is required",
		},
		{
			name: "missing artist",
			song: models.Song{
				ID:        "song-123",
				Title:     "Test Song",
				S3Key:     "songs/test-song.mp3",
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "song artist is required",
		},
		{
			name: "missing S3Key",
			song: models.Song{
				ID:        "song-123",
				Title:     "Test Song",
				Artist:    "Test Artist",
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "S3 key is required",
		},
		{
			name: "invalid year",
			song: models.Song{
				ID:        "song-123",
				Title:     "Test Song",
				Artist:    "Test Artist",
				Year:      1500,
				S3Key:     "songs/test-song.mp3",
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid year: must be between 1800 and current year",
		},
		{
			name: "missing created_at",
			song: models.Song{
				ID:     "song-123",
				Title:  "Test Song",
				Artist: "Test Artist",
				Year:   2023,
				S3Key:  "songs/test-song.mp3",
			},
			wantErr: true,
			errMsg:  "created_at timestamp is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearTestData(t, db)

			err := repo.SaveSong(tt.song)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)

				// Verify song was saved
				saved, err := repo.FindById(tt.song.ID)
				require.NoError(t, err)
				require.NotNil(t, saved)

				assert.Equal(t, tt.song.ID, saved.ID)
				assert.Equal(t, tt.song.Title, saved.Title)
				assert.Equal(t, tt.song.Artist, saved.Artist)
				assert.Equal(t, tt.song.Album, saved.Album)
				assert.Equal(t, tt.song.Year, saved.Year)
				assert.Equal(t, tt.song.S3Key, saved.S3Key)
				assert.Equal(t, tt.song.Fingerprint, saved.Fingerprint)
				assert.WithinDuration(t, tt.song.CreatedAt, saved.CreatedAt, time.Minute)
			}
		})
	}
}

func TestSongRepo_FindById(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewSongRepo(db)

	// Setup test data
	testSong := models.Song{
		ID:          "123", // Match fingerprint SongID
		Title:       "Test Song",
		Artist:      "Test Artist",
		Album:       "Test Album",
		Year:        2023,
		S3Key:       "songs/test-song.mp3",
		Fingerprint: []byte("test-fingerprint"),
		CreatedAt:   time.Now(),
	}

	err := repo.SaveSong(testSong)
	require.NoError(t, err)

	tests := []struct {
		name     string
		id       string
		wantSong *models.Song
		wantErr  bool
	}{
		{
			name:     "existing song",
			id:       "123",
			wantSong: &testSong,
			wantErr:  false,
		},
		{
			name:     "non-existing song",
			id:       "non-existing",
			wantSong: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindById(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantSong == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantSong.ID, result.ID)
				assert.Equal(t, tt.wantSong.Title, result.Title)
				assert.Equal(t, tt.wantSong.Artist, result.Artist)
				assert.Equal(t, tt.wantSong.Album, result.Album)
				assert.Equal(t, tt.wantSong.Year, result.Year)
				assert.Equal(t, tt.wantSong.S3Key, result.S3Key)
				assert.Equal(t, tt.wantSong.Fingerprint, result.Fingerprint)
			}
		})
	}
}

func TestSongRepo_FindByFingerprint(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Setup test data
	testSong := models.Song{
		ID:          "123", // Match fingerprint SongID
		Title:       "Test Song",
		Artist:      "Test Artist",
		Album:       "Test Album",
		Year:        2023,
		S3Key:       "songs/test-song.mp3",
		Fingerprint: []byte("test-fingerprint"),
		CreatedAt:   time.Now(),
	}

	err := songRepo.SaveSong(testSong)
	require.NoError(t, err)

	testFingerprint := models.Fingerprint{
		SongID:     123, // This should match testSong.ID "song-123"
		Hash:       12345,
		TimeOffset: 1000,
	}

	err = fingerprintRepo.SaveFingerprint(testFingerprint)
	require.NoError(t, err)

	tests := []struct {
		name     string
		hash     string
		wantSong *models.Song
		wantErr  bool
	}{
		{
			name:     "existing fingerprint hash",
			hash:     "12345",
			wantSong: &testSong,
			wantErr:  false,
		},
		{
			name:     "non-existing fingerprint hash",
			hash:     "99999",
			wantSong: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := songRepo.FindByFingerprint(tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantSong == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantSong.ID, result.ID)
				assert.Equal(t, tt.wantSong.Title, result.Title)
				assert.Equal(t, tt.wantSong.Artist, result.Artist)
			}
		})
	}
}

func TestSongRepo_DuplicateID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewSongRepo(db)

	song1 := models.Song{
		ID:        "song-123",
		Title:     "Test Song 1",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song-1.mp3",
		CreatedAt: time.Now(),
	}

	song2 := models.Song{
		ID:        "song-123", // Same ID
		Title:     "Test Song 2",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song-2.mp3",
		CreatedAt: time.Now(),
	}

	// First save should succeed
	err := repo.SaveSong(song1)
	assert.NoError(t, err)

	// Second save with same ID should fail
	err = repo.SaveSong(song2)
	assert.Error(t, err)
}
