package server

import (
	"github.com/owenhochwald/harmonia/internal/config"
	"github.com/rs/zerolog"
)

// TODO: add model properties
type Application struct {
	Logger zerolog.Logger
	Config config.Config
}
