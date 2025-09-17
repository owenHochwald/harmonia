package repo

import (
	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockSongRepo struct {
	mock.Mock
}

func (m *MockSongRepo) SaveSong(song models.Song) error {
	args := m.Called(song)
	return args.Error(0)
}

func (m *MockSongRepo) FindById(id string) (*models.Song, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Song), args.Error(1)
}

func (m *MockSongRepo) FindByFingerprint(hash string) (*models.Song, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Song), args.Error(1)
}

type MockFingerprintRepo struct {
	mock.Mock
}

func (m *MockFingerprintRepo) SaveFingerprint(fingerprint models.Fingerprint) error {
	args := m.Called(fingerprint)
	return args.Error(0)
}

func (m *MockFingerprintRepo) FindByHash(hash string) (*models.Fingerprint, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Fingerprint), args.Error(1)
}

func (m *MockFingerprintRepo) FindById(id int64) (*models.Fingerprint, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Fingerprint), args.Error(1)
}

func (m *MockFingerprintRepo) FindBySongId(songId string) (*models.Fingerprint, error) {
	args := m.Called(songId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Fingerprint), args.Error(1)
}

func NewMockSongRepo() *MockSongRepo {
	return &MockSongRepo{}
}

// NewMockFingerprintRepo creates a new mock fingerprint repository
func NewMockFingerprintRepo() *MockFingerprintRepo {
	return &MockFingerprintRepo{}
}
