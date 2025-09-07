package main

import (
	"github.com/gin-gonic/gin"
	config2 "github.com/owenhochwald/harmonia/internal/config"
	"github.com/owenhochwald/harmonia/internal/server"
)

func main() {
	// TODO: add full setup for config struct
	config := config2.LoadConfig()

	// TODO: add full setup for app
	app := server.Application{}

	r := gin.Default()

	server.SetupRoutes(r, &app)

	r.Run(":" + config.Port)
}
