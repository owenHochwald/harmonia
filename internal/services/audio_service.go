package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/youpy/go-wav"
	"github.com/zaf/resample"
)

type AudioServiceInterface interface {
	ValidateFile(r *bytes.Reader) (error, int)
	Process(raw []byte) (*AudioData, error)
	ReadWAVProperties(r *bytes.Reader) (*AudioMetadata, error)
}

type AudioService struct {
	Data *AudioData
}

func NewAudioService() AudioServiceInterface {
	return &AudioService{
		Data: &AudioData{},
	}
}

type AudioData struct {
	Metadata AudioMetadata
	Audio    ProcessedAudio
}

type ProcessedAudio struct {
	Samples       []float64     `json:"-"` // Don't serialize raw samples
	SampleRate    uint32        `json:"sample_rate"`
	Duration      time.Duration `json:"duration"`
	NumSamples    int           `json:"num_samples"`
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
	TotalSamples     int64   `json:"total_samples"`
	SampleRate       uint32  `json:"sample_rate"`
}

func (a *AudioService) Process(raw []byte) (*AudioData, error) {
	// TODO: Implement me!
	a.Data.Metadata = *a.ExtractMetadata(raw)

	// TODO: error handling
	mono, _ := a.ConvertToMono(raw)

	a.Resample(mono, float64(16000))

	filtered := a.FilterFrequencies(mono)

	normalized := a.Normalize(filtered, a.Data.Metadata.SampleRate, 16000)

	spectrogram := a.Spectrogram(normalized, 2048, 512)

	a.ExtractMFCCs(spectrogram, 20)
	a.ExtractSpectralFeatures(spectrogram)

	return a.Data, nil
}

func (a *AudioService) ExtractMetadata(data []byte) *AudioMetadata {
	// TODO: Implement me
	return nil
}

func (a *AudioService) ValidateFile(r *bytes.Reader) (error, int) {
	if r.Len() == 0 {
		return fmt.Errorf("empty file"), http.StatusBadRequest
	}
	const maxFileSize = 10 * 1024 * 1024
	if int64(r.Len()) > maxFileSize {
		return fmt.Errorf("file is too large"), http.StatusBadRequest
	}
	wavReader := wav.NewReader(r)
	format, err := wavReader.Format()
	if err != nil {
		return fmt.Errorf("error reading WAV format: %w", err), http.StatusInternalServerError
	} else if format.AudioFormat != wav.AudioFormatPCM {
		return fmt.Errorf("unsupported audio format: %d", format.AudioFormat), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func (a *AudioService) GetTotalSamples(data []byte) (int, error) {
	if len(data) < 44 {
		return 0, fmt.Errorf("file is too small")
	}

	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)

	format, err := wavReader.Format()

	if err != nil {
		return 0, fmt.Errorf("failed to read format: %w", err)
	}

	dataChunkSize := binary.LittleEndian.Uint32(data[40:44])

	bytesPerSample := format.BitsPerSample / 8
	bytesPerFrame := bytesPerSample * format.NumChannels

	totalFrames := int(dataChunkSize) / int(bytesPerFrame)

	return totalFrames, nil
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

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes form reader: %v", err)
	}

	totalSamples, err := a.GetTotalSamples(data)

	metaData := &AudioMetadata{
		OriginalFormat:   "WAV",
		OriginalChannels: format.NumChannels,
		OriginalRate:     format.SampleRate,
		OriginalBits:     format.BitsPerSample,
		Duration:         seconds,
		FileSize:         size,
		TotalSamples:     int64(totalSamples),
		SampleRate:       format.SampleRate,
	}

	return metaData, nil
}
func (a *AudioService) ConvertToMono(data []byte) ([]byte, error) {
	metadata, err := a.ReadWAVProperties(bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	if metadata.OriginalChannels == 1 {
		return data, nil
	}

	totalFrames, _ := a.GetTotalSamples(data)

	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, _ := wavReader.Format()

	//var monoSamples []wav.Sample
	monoSamples := make([]wav.Sample, 0, totalFrames)

	for {
		samples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, sample := range samples {
			sum := int64(0)
			for ch := uint(0); ch < uint(format.NumChannels); ch++ {
				sum += int64(wavReader.IntValue(sample, ch))
			}
			monoValue := int(sum / int64(format.NumChannels))

			monoSamples = append(monoSamples, wav.Sample{
				Values: [2]int{monoValue, monoValue},
			})
		}
	}

	var outputBuffer bytes.Buffer

	writer := wav.NewWriter(&outputBuffer, uint32(len(monoSamples)), 1, format.SampleRate, format.BitsPerSample)

	if err := writer.WriteSamples(monoSamples); err != nil {
		return nil, fmt.Errorf("failed to write samples: %w", err)
	}

	return outputBuffer.Bytes(), nil
}

func (a *AudioService) Resample(data []byte, targetSampleRate float64) ([]byte, error) {
	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read format: %w", err)
	}

	if float64(format.SampleRate) == targetSampleRate {
		return data, nil
	}

	pcmData := data[44:]

	inputBuffer := bytes.NewBuffer(pcmData)
	outputBuffer := &bytes.Buffer{}

	resampler, err := resample.New(outputBuffer, float64(format.SampleRate), targetSampleRate, int(format.NumChannels), resample.I16, resample.I32)
	if err != nil {
		return nil, fmt.Errorf("error creating resampler: %w", err)
	}

	_, err = io.Copy(resampler, inputBuffer)
	if err != nil {
		return nil, fmt.Errorf("error resampling data: %w", err)
	}

	if err := resampler.Close(); err != nil {
		return nil, fmt.Errorf("error closing resampler: %w", err)
	}

	resampledData := outputBuffer.Bytes()
	return resampledData, nil
}
func (a *AudioService) ExtractMFCCs(spectrogram any, i int) interface{} {
	return nil
}

func (a *AudioService) Normalize(mono []byte, rate uint32, i int) interface{} {
	return nil
}

func (a *AudioService) ExtractSpectralFeatures(spectrogram any) interface{} {
	return nil
}

func (a *AudioService) Spectrogram(normalized interface{}, i int, i2 int) interface{} {
	return nil
}

func (a *AudioService) FilterFrequencies(FilterFrequencies interface{}) []byte {
	return nil
}
