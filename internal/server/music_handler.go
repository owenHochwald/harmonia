package server

import (
	"bytes"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/owenhochwald/harmonia/internal/services"
)

func (app *Application) handleTestWaveUpload(c *gin.Context) {
	file, err := os.Open("/Users/owenhochwald/Documents/code/personal/backend/go/harmonia/public/audios/sample-12s.wav")
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to open file")
		c.JSON(500, gin.H{"error": "Failed to open audio file"})
		return
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read file")
		c.JSON(500, gin.H{"error": "Failed to read audio file"})
		return
	}

	reader := bytes.NewReader(data)

	audioService := &services.AudioService{}
	metaData, err := audioService.ReadWAVProperties(reader)

	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read WAV properties")
		c.JSON(500, gin.H{"error": "Failed to read WAV properties"})
		return
	}

	c.JSON(200, metaData)
}
