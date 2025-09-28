package services

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/youpy/go-wav"
)

type AudioService struct{}

type ProcessedAudio struct {
	Samples       []float64     `json:"-"` // Don't serialize raw samples
	SampleRate    uint32        `json:"sample_rate"`
	Duration      time.Duration `json:"duration"`
	NumSamples    int           `json:"num_samples"`
	Channels      uint16        `json:"original_channels"`
	BitsPerSample uint16        `json:"original_bits_per_sample"`
	ProcessedAt   time.Time     `json:"processed_at"`
}

type AudioMetadata struct {
	OriginalFormat   string  `json:"original_format"`
	OriginalChannels uint16  `json:"original_channels"`
	OriginalRate     uint32  `json:"original_sample_rate"`
	OriginalBits     uint16  `json:"original_bits_per_sample"`
	Duration         float64 `json:"duration"`
	FileSize         int64   `json:"file_size_bytes"`
}

func (a *AudioService) Process(r io.Reader, originalSize int64) (*ProcessedAudio, error) {
	// TODO: Implement me!
	return nil, nil
}

func (a *AudioService) ReadWAVProperties(r *bytes.Reader) (*AudioMetadata, error) {
	wavReader := wav.NewReader(r)

	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("error reading WAV format: %w", err)
	}

	if format.AudioFormat != wav.AudioFormatPCM {
		return nil, fmt.Errorf("unsupported audio format: %d", format.AudioFormat)
	}

	duration, err := wavReader.Duration()
	seconds := duration.Seconds()
	if err != nil {
		return nil, fmt.Errorf("error calculating WAV duration: %w", err)
	}

	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("error determining file size: %w", err)
	}

	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("error resetting reader position: %w", err)
	}

	metaData := &AudioMetadata{
		OriginalFormat:   "WAV",
		OriginalChannels: format.NumChannels,
		OriginalRate:     format.SampleRate,
		OriginalBits:     format.BitsPerSample,
		Duration:         seconds,
		FileSize:         size,
	}

	return metaData, nil
}
