package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/owenhochwald/harmonia/internal/config"
	"github.com/owenhochwald/harmonia/internal/server"
	"github.com/owenhochwald/harmonia/pkg/logger"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	app := server.Application{
		Logger: logger.NewLogger(os.Getenv("ENVIRONMENT")),
		Config: config.NewConfig(),
	}

	r := gin.Default()

	server.SetupRoutes(r, &app)

	r.Run(":" + app.Config.Port)
}
