package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/owenhochwald/harmonia/internal/config"
	"github.com/owenhochwald/harmonia/internal/server"
	"github.com/owenhochwald/harmonia/pkg/logger"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	log := logger.NewLogger(os.Getenv("ENVIRONMENT"))
	cfg := config.NewConfig()

	db, err := ConnectDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to database")
	}

	app := server.Application{
		Logger: log,
		Config: cfg,
		DB:     db,
	}

	r := gin.Default()

	server.SetupRoutes(r, &app)

	r.Run(":" + app.Config.Port)
}

func ConnectDB(cfg config.Config) (*sql.DB, error) {
	var err error
	db, err := sql.Open("postgres", cfg.DBURL)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxOpenConns(25)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
