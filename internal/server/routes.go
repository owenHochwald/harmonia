package server

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, app *Application) {
	r.GET("/health", app.HealthHandler.Check)
	r.GET("/test-wave-upload", app.MusicHandler.handleTestWaveUpload)
	r.POST("/music/upload", app.MusicHandler.handleAudioUpload)
}
