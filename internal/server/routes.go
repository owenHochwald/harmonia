package server

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, app *Application) {
	r.GET("/health", app.HealthHandler.Check)
	r.GET("/test-wave-upload", app.MusicHandler.handleTestWaveUpload)

	r.POST("/api/upload", app.MusicHandler.handleAudioUpload)
	r.GET("/api/songs", app.MusicHandler.handleGetSongs)
	r.GET("/api/songs:id", app.MusicHandler.handleGetASong)
}
