package server

import (
	"github.com/owenhochwald/harmonia/internal/config"
	"github.com/rs/zerolog"
)

type Application struct {
	logger zerolog.Logger
	config config.Config
}
