package repo

import (
	"errors"
	"testing"
	"time"

	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Example of how to use the mocks in unit tests
func TestMockSongRepo_Usage(t *testing.T) {
	mockRepo := NewMockSongRepo()

	testSong := models.Song{
		ID:        "song-123",
		Title:     "Test Song",
		Artist:    "Test Artist",
		S3Key:     "songs/test.mp3",
		CreatedAt: time.Now(),
	}

	t.Run("SaveSong success", func(t *testing.T) {
		// Setup expectation
		mockRepo.On("SaveSong", testSong).Return(nil).Once()

		// Call the method
		err := mockRepo.SaveSong(testSong)

		// Assertions
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("SaveSong failure", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		expectedErr := errors.New("database error")
		mockRepo.On("SaveSong", testSong).Return(expectedErr).Once()

		// Call the method
		err := mockRepo.SaveSong(testSong)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindById success", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindById", "song-123").Return(&testSong, nil).Once()

		// Call the method
		result, err := mockRepo.FindById("song-123")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testSong.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindById not found", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindById", "non-existing").Return(nil, nil).Once()

		// Call the method
		result, err := mockRepo.FindById("non-existing")

		// Assertions
		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindByFingerprint success", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindByFingerprint", "12345").Return(&testSong, nil).Once()

		// Call the method
		result, err := mockRepo.FindByFingerprint("12345")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testSong.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockFingerprintRepo_Usage(t *testing.T) {
	mockRepo := NewMockFingerprintRepo()

	testFingerprint := models.Fingerprint{
		ID:         1,
		SongID:     123,
		Hash:       12345,
		TimeOffset: 1000,
	}

	t.Run("SaveFingerprint success", func(t *testing.T) {
		mockRepo.On("SaveFingerprint", testFingerprint).Return(nil).Once()

		err := mockRepo.SaveFingerprint(testFingerprint)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindByHash success", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindByHash", "12345").Return(&testFingerprint, nil).Once()

		result, err := mockRepo.FindByHash("12345")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testFingerprint.Hash, result.Hash)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindById success", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindById", int64(1)).Return(&testFingerprint, nil).Once()

		result, err := mockRepo.FindById(1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testFingerprint.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindBySongId success", func(t *testing.T) {
		// Reset mock for new test
		mockRepo.ExpectedCalls = nil

		mockRepo.On("FindBySongId", "123").Return(&testFingerprint, nil).Once()

		result, err := mockRepo.FindBySongId("123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testFingerprint.SongID, result.SongID)
		mockRepo.AssertExpectations(t)
	})
}

// Example of using mocks with multiple expectations
func TestMockRepo_MultipleExpectations(t *testing.T) {
	mockSongRepo := NewMockSongRepo()
	mockFingerprintRepo := NewMockFingerprintRepo()

	song := models.Song{
		ID:        "song-456",
		Title:     "Complex Test",
		Artist:    "Test Artist",
		S3Key:     "songs/complex.mp3",
		CreatedAt: time.Now(),
	}

	fingerprint := models.Fingerprint{
		SongID:     456,
		Hash:       67890,
		TimeOffset: 2000,
	}

	// Setup multiple expectations
	mockSongRepo.On("SaveSong", song).Return(nil).Once()
	mockFingerprintRepo.On("SaveFingerprint", fingerprint).Return(nil).Once()
	mockSongRepo.On("FindById", "song-456").Return(&song, nil).Once()
	mockFingerprintRepo.On("FindBySongId", "456").Return(&fingerprint, nil).Once()

	// Simulate a workflow
	err := mockSongRepo.SaveSong(song)
	assert.NoError(t, err)

	err = mockFingerprintRepo.SaveFingerprint(fingerprint)
	assert.NoError(t, err)

	foundSong, err := mockSongRepo.FindById("song-456")
	assert.NoError(t, err)
	assert.Equal(t, song.ID, foundSong.ID)

	foundFingerprint, err := mockFingerprintRepo.FindBySongId("456")
	assert.NoError(t, err)
	assert.Equal(t, fingerprint.SongID, foundFingerprint.SongID)

	// Verify all expectations were met
	mockSongRepo.AssertExpectations(t)
	mockFingerprintRepo.AssertExpectations(t)
}

// Example of testing with mock that uses argument matchers
func TestMockRepo_ArgumentMatchers(t *testing.T) {
	mockRepo := NewMockSongRepo()

	// Use argument matchers for more flexible expectations
	mockRepo.On("FindById", mock.AnythingOfType("string")).Return(nil, nil)

	// This will match any string argument
	result, err := mockRepo.FindById("any-id")
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = mockRepo.FindById("another-id")
	assert.NoError(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
