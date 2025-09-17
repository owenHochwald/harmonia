package repo

import (
	"strconv"
	"testing"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFingerprintRepo_SaveFingerprint(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Setup test song first
	testSong := models.Song{
		ID:        "123", // Use numeric string to match fingerprint SongID
		Title:     "Test Song",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song.mp3",
		CreatedAt: time.Now(),
	}
	err := songRepo.SaveSong(testSong)
	require.NoError(t, err)

	tests := []struct {
		name        string
		fingerprint models.Fingerprint
		wantErr     bool
	}{
		{
			name: "valid fingerprint",
			fingerprint: models.Fingerprint{
				SongID:     123, // This corresponds to song-123 when cast to string
				Hash:       12345,
				TimeOffset: 1000,
			},
			wantErr: false,
		},
		{
			name: "another valid fingerprint",
			fingerprint: models.Fingerprint{
				SongID:     123,
				Hash:       67890,
				TimeOffset: 2000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fingerprintRepo.SaveFingerprint(tt.fingerprint)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFingerprintRepo_FindById(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Setup test song first
	testSong := models.Song{
		ID:        "123", // Use numeric string to match fingerprint SongID
		Title:     "Test Song",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song.mp3",
		CreatedAt: time.Now(),
	}
	err := songRepo.SaveSong(testSong)
	require.NoError(t, err)

	// Setup test fingerprint
	testFingerprint := models.Fingerprint{
		SongID:     123, // This will need to match the song's ID in string form
		Hash:       12345,
		TimeOffset: 1000,
	}

	err = fingerprintRepo.SaveFingerprint(testFingerprint)
	require.NoError(t, err)

	// Get the saved fingerprint to get its ID
	savedFingerprint, err := fingerprintRepo.FindByHash("12345")
	require.NoError(t, err)
	require.NotNil(t, savedFingerprint)

	tests := []struct {
		name            string
		id              int64
		wantFingerprint *models.Fingerprint
		wantErr         bool
	}{
		{
			name:            "existing fingerprint",
			id:              savedFingerprint.ID,
			wantFingerprint: savedFingerprint,
			wantErr:         false,
		},
		{
			name:            "non-existing fingerprint",
			id:              99999,
			wantFingerprint: nil,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fingerprintRepo.FindById(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantFingerprint == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantFingerprint.ID, result.ID)
				assert.Equal(t, tt.wantFingerprint.SongID, result.SongID)
				assert.Equal(t, tt.wantFingerprint.Hash, result.Hash)
				assert.Equal(t, tt.wantFingerprint.TimeOffset, result.TimeOffset)
			}
		})
	}
}

func TestFingerprintRepo_FindByHash(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Setup test song first
	testSong := models.Song{
		ID:        "123", // Use numeric string to match fingerprint SongID
		Title:     "Test Song",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song.mp3",
		CreatedAt: time.Now(),
	}
	err := songRepo.SaveSong(testSong)
	require.NoError(t, err)

	// Setup test fingerprints
	fingerprint1 := models.Fingerprint{
		SongID:     123,
		Hash:       12345,
		TimeOffset: 1000,
	}

	fingerprint2 := models.Fingerprint{
		SongID:     123,
		Hash:       67890,
		TimeOffset: 2000,
	}

	err = fingerprintRepo.SaveFingerprint(fingerprint1)
	require.NoError(t, err)

	err = fingerprintRepo.SaveFingerprint(fingerprint2)
	require.NoError(t, err)

	tests := []struct {
		name            string
		hash            string
		wantFingerprint bool
		expectedHash    uint32
		wantErr         bool
	}{
		{
			name:            "existing hash 1",
			hash:            "12345",
			wantFingerprint: true,
			expectedHash:    12345,
			wantErr:         false,
		},
		{
			name:            "existing hash 2",
			hash:            "67890",
			wantFingerprint: true,
			expectedHash:    67890,
			wantErr:         false,
		},
		{
			name:            "non-existing hash",
			hash:            "99999",
			wantFingerprint: false,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fingerprintRepo.FindByHash(tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.wantFingerprint {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedHash, result.Hash)
				assert.Equal(t, int64(123), result.SongID)
			}
		})
	}
}

func TestFingerprintRepo_FindBySongId(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Setup test songs
	testSong1 := models.Song{
		ID:        "456", // Use numeric string to match fingerprint SongID
		Title:     "Test Song 1",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song-1.mp3",
		CreatedAt: time.Now(),
	}

	testSong2 := models.Song{
		ID:        "789",
		Title:     "Test Song 2",
		Artist:    "Test Artist",
		Year:      2023,
		S3Key:     "songs/test-song-2.mp3",
		CreatedAt: time.Now(),
	}

	err := songRepo.SaveSong(testSong1)
	require.NoError(t, err)

	err = songRepo.SaveSong(testSong2)
	require.NoError(t, err)

	// Setup test fingerprints for each song
	fingerprint1 := models.Fingerprint{
		SongID:     456, // Matches testSong1.ID
		Hash:       12345,
		TimeOffset: 1000,
	}

	fingerprint2 := models.Fingerprint{
		SongID:     789, // Matches testSong2.ID
		Hash:       67890,
		TimeOffset: 2000,
	}

	err = fingerprintRepo.SaveFingerprint(fingerprint1)
	require.NoError(t, err)

	err = fingerprintRepo.SaveFingerprint(fingerprint2)
	require.NoError(t, err)

	tests := []struct {
		name            string
		songId          string
		wantFingerprint bool
		expectedSongID  int64
		wantErr         bool
	}{
		{
			name:            "existing song with fingerprints",
			songId:          "456",
			wantFingerprint: true,
			expectedSongID:  456,
			wantErr:         false,
		},
		{
			name:            "another existing song",
			songId:          "789",
			wantFingerprint: true,
			expectedSongID:  789,
			wantErr:         false,
		},
		{
			name:            "non-existing song",
			songId:          "999",
			wantFingerprint: false,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fingerprintRepo.FindBySongId(tt.songId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.wantFingerprint {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedSongID, result.SongID)
			}
		})
	}
}

func TestFingerprintRepo_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	songRepo := NewSongRepo(db)
	fingerprintRepo := NewFingerprintRepo(db)

	// Create a complete workflow test
	testSong := models.Song{
		ID:        "999", // Use numeric string for foreign key compatibility
		Title:     "Integration Test Song",
		Artist:    "Test Artist",
		Album:     "Test Album",
		Year:      2023,
		S3Key:     "songs/integration-test.mp3",
		CreatedAt: time.Now(),
	}

	// 1. Save song
	err := songRepo.SaveSong(testSong)
	require.NoError(t, err)

	// 2. Create multiple fingerprints for the song (using numeric song ID)
	songIDInt, _ := strconv.ParseInt(testSong.ID, 10, 64)
	fingerprints := []models.Fingerprint{
		{SongID: songIDInt, Hash: 11111, TimeOffset: 1000},
		{SongID: songIDInt, Hash: 22222, TimeOffset: 2000},
		{SongID: songIDInt, Hash: 33333, TimeOffset: 3000},
	}

	for _, fp := range fingerprints {
		err := fingerprintRepo.SaveFingerprint(fp)
		require.NoError(t, err)
	}

	// 3. Test finding song by any fingerprint hash
	for _, fp := range fingerprints {
		// Find fingerprint by hash
		foundFp, err := fingerprintRepo.FindByHash(strconv.Itoa(int(fp.Hash)))
		require.NoError(t, err)
		require.NotNil(t, foundFp)
		assert.Equal(t, fp.Hash, foundFp.Hash)

		// Find song by fingerprint hash
		foundSong, err := songRepo.FindByFingerprint(strconv.Itoa(int(fp.Hash)))
		require.NoError(t, err)
		require.NotNil(t, foundSong)
		assert.Equal(t, testSong.ID, foundSong.ID)
		assert.Equal(t, testSong.Title, foundSong.Title)
	}

	// 4. Test finding fingerprint by song ID
	foundFp, err := fingerprintRepo.FindBySongId(testSong.ID)
	require.NoError(t, err)
	require.NotNil(t, foundFp)
	assert.Equal(t, songIDInt, foundFp.SongID)
}
