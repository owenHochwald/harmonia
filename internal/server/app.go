package server

import (
	"database/sql"

	"github.com/owenhochwald/harmonia/internal/config"
	"github.com/owenhochwald/harmonia/internal/repo"
	"github.com/owenhochwald/harmonia/internal/services"
	"github.com/owenhochwald/harmonia/internal/storage"
	"github.com/rs/zerolog"
)

type Application struct {
	Logger zerolog.Logger
	Config config.Config
	DB     *sql.DB

	Storage storage.Storage

	SongRepo        repo.SongRepo
	FingerprintRepo repo.FingerprintRepo

	AudioService       services.AudioServiceInterface
	MusicService       services.MusicServiceInterface
	FingerprintService services.FingerprintServiceInterface

	MusicHandler  *MusicHandler
	HealthHandler *HealthHandler
}

func NewApplication(cfg config.Config, log zerolog.Logger, db *sql.DB) (*Application, error) {
	app := &Application{
		Logger: log,
		Config: cfg,
		DB:     db,
	}

	// TODO: add concrete implementation for S3 storage
	//app.Storage =

	if err := app.initRepos(); err != nil {
		return nil, err
	}
	if err := app.initServices(); err != nil {
		return nil, err
	}
	if err := app.initHandlers(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) initRepos() error {
	app.SongRepo = repo.NewSongRepo(app.DB)
	app.FingerprintRepo = repo.NewFingerprintRepo(app.DB)

	return nil
}

func (app *Application) initServices() error {
	app.MusicService = services.NewMusicService(app.Storage, app.SongRepo, app.AudioService, app.FingerprintService)
	app.AudioService = services.NewAudioService()
	app.FingerprintService = services.NewFingerprintService(app.FingerprintRepo)

	return nil
}

func (app *Application) initHandlers() error {
	app.MusicHandler = NewMusicHandler(app.AudioService, app.MusicService, app.SongRepo)
	app.HealthHandler = NewHealthHandler()

	return nil
}
