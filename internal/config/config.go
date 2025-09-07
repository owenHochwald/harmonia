package config

import (
	"github.com/rs/zerolog"
)

type Config struct {
	Logger    zerolog.Logger
	Env       string
	Port      string
	DBURL     string
	S3Bucket  string
	AWSRegion string
}

func LoadConfig() Config {
	// TODO: Implement load ENV VAR functionality
	return Config{}
}
