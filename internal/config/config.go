package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/owenhochwald/harmonia/pkg/logger"
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

func NewConfig() Config {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	env := os.Getenv("ENVIRONMENT")
	port := os.Getenv("PORT")
	db_url := os.Getenv("DB_URL")
	region := os.Getenv("AWS_REGION")
	s3_bucket := os.Getenv("S3_BUCKET")

	logger := logger.NewLogger(env)

	return Config{
		Logger:    logger,
		Env:       env,
		Port:      port,
		DBURL:     db_url,
		S3Bucket:  s3_bucket,
		AWSRegion: region,
	}

}
