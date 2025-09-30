package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/owenhochwald/harmonia/internal/services"
)

func (app *Application) handleAudioUpload(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to get file"})
	}
	defer file.Close()

	audioBytes, err := io.ReadAll(file)

	if err != nil || len(audioBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "please provide a valid file"})
	}

	if err = c.Request.Body.Close(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error reading and closing the request body"})
	}
	app.Logger.Info().Int("size_bytes", len(audioBytes)).Msg("audio file received")

	audioService := &services.AudioService{}

	if err, code := audioService.ValidateFile(bytes.NewReader(audioBytes)); err != nil {
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	metaData, err := audioService.ReadWAVProperties(bytes.NewReader(audioBytes))

	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read WAV properties")
		c.JSON(500, gin.H{"error": "Failed to read WAV properties"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "metadata": metaData})
}

func (app *Application) handleTestWaveUpload(c *gin.Context) {
	params := c.DefaultQuery("test", "properties")
	download := c.DefaultQuery("download", "false")

	switch params {
	case "properties":
		app.testAudioProperties(c)
		return
	case "mono":
		if download == "true" {
			app.downloadMonoConversion(c)
		} else {
			app.testMonoConversion(c)
		}
		return
	default:
		c.JSON(400, gin.H{"error": "Invalid test parameter, please use 'properties' or 'mono'"})
	}

	app.testAudioProperties(c)
}

func (app *Application) testAudioProperties(c *gin.Context) {
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

func (app *Application) testMonoConversion(c *gin.Context) {
	audioFile := "/Users/owenhochwald/Documents/code/personal/backend/go/harmonia/public/audios/sample-12s.wav"

	data, err := app.readAudioFile(audioFile)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read audio file")
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	audioService := &services.AudioService{}

	originalMetadata, err := audioService.ReadWAVProperties(bytes.NewReader(data))
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read original properties")
		c.JSON(500, gin.H{"error": "Failed to read original properties"})
		return
	}

	monoData, err := audioService.ConvertToMono(data)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to convert to mono")
		c.JSON(500, gin.H{"error": "Failed to convert to mono", "details": err.Error()})
		return
	}

	convertedMetadata, err := audioService.ReadWAVProperties(bytes.NewReader(monoData))
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read converted properties")
		c.JSON(500, gin.H{"error": "Failed to read converted properties"})
		return
	}

	app.Logger.Info().
		Uint16("original_channels", originalMetadata.OriginalChannels).
		Uint16("converted_channels", convertedMetadata.OriginalChannels).
		Int("original_size", len(data)).
		Int("converted_size", len(monoData)).
		Msg("Successfully converted to mono")

	c.JSON(200, gin.H{
		"test":      "mono_conversion",
		"success":   true,
		"original":  originalMetadata,
		"converted": convertedMetadata,
		"size_change": gin.H{
			"original_bytes":    len(data),
			"converted_bytes":   len(monoData),
			"reduction_percent": float64(len(data)-len(monoData)) / float64(len(data)) * 100,
		},
		"message":      "Mono conversion completed successfully",
		"download_url": "/test/audio?test=mono&download=true", // Tell user how to download
	})
}

func (app *Application) downloadMonoConversion(c *gin.Context) {
	audioFile := "/Users/owenhochwald/Documents/code/personal/backend/go/harmonia/public/audios/sample-12s.wav"

	data, err := app.readAudioFile(audioFile)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to read audio file")
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	audioService := &services.AudioService{}

	monoData, err := audioService.ConvertToMono(data)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Failed to convert to mono")
		c.JSON(500, gin.H{"error": "Failed to convert to mono", "details": err.Error()})
		return
	}

	c.Data(200, "audio/wav", monoData)
}

func (app *Application) readAudioFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file not found: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			app.Logger.Error().Err(err).Msg("Failed to close file")
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
