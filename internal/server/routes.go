package server

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, app *Application) {
	r.GET("/health", app.HealthHandler)
	r.GET("/test-wave-upload", app.handleTestWaveUpload)
}
