package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/mjibson/go-dsp/fft"
	"github.com/youpy/go-wav"
	"github.com/zeozeozeo/gomplerate"
)

type AudioServiceInterface interface {
	ValidateFile(r *bytes.Reader) (error, int)
	Process(raw []byte) (*AudioData, error)
	ReadWAVProperties(r *bytes.Reader) (*AudioMetadata, error)
	ConvertToMono(data []byte) ([]byte, error)
	Resample(data []byte, targetSampleRate uint32) ([]byte, error)
	Normalize(data []byte) ([]byte, error)
	Spectrogram(data []byte, windowSize, hopSize int) (*Spectrogram, error)
	GetTotalSamples(data []byte) (int, error)
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
	Samples       []float64     `json:"-"`
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

type Spectrogram struct {
	Data          [][]float64
	FrequencyBins []float64
	TimeFrames    []float64
	SampleRate    uint32
}

func (a *AudioService) Process(raw []byte) (*AudioData, error) {
	metadata, err := a.ReadWAVProperties(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	a.Data.Metadata = *metadata

	mono, err := a.ConvertToMono(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to mono: %w", err)
	}

	resampled, err := a.Resample(mono, 16000)
	if err != nil {
		return nil, fmt.Errorf("failed to resample: %w", err)
	}

	normalized, err := a.Normalize(resampled)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize: %w", err)
	}

	spectrogram, err := a.Spectrogram(normalized, 2048, 512)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spectrogram: %w", err)
	}

	// TODO: Extract features from spectrogram
	_ = spectrogram

	return a.Data, nil
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
	}
	if format.AudioFormat != wav.AudioFormatPCM {
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
		return nil, fmt.Errorf("error reading bytes from reader: %w", err)
	}

	totalSamples, err := a.GetTotalSamples(data)
	if err != nil {
		return nil, fmt.Errorf("error getting total samples: %w", err)
	}

	metaData := &AudioMetadata{
		OriginalFormat:   "WAV",
		OriginalChannels: format.NumChannels,
		OriginalRate:     format.SampleRate,
		OriginalBits:     format.BitsPerSample,
		Duration:         duration.Seconds(),
		FileSize:         size,
		TotalSamples:     int64(totalSamples),
		SampleRate:       format.SampleRate,
	}

	return metaData, nil
}

func (a *AudioService) ConvertToMono(data []byte) ([]byte, error) {
	metadata, err := a.ReadWAVProperties(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to read properties: %w", err)
	}

	if metadata.OriginalChannels == 1 {
		return data, nil
	}

	totalFrames, err := a.GetTotalSamples(data)
	if err != nil {
		return nil, fmt.Errorf("failed to get total samples: %w", err)
	}

	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read format: %w", err)
	}

	monoSamples := make([]wav.Sample, 0, totalFrames)

	for {
		samples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read samples: %w", err)
		}

		for _, sample := range samples {
			sum := int64(0)
			for ch := uint(0); ch < uint(format.NumChannels); ch++ {
				sum += int64(wavReader.IntValue(sample, ch))
			}
			monoValue := int(sum / int64(format.NumChannels))

			monoSamples = append(monoSamples, wav.Sample{
				Values: [2]int{monoValue, 0},
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

func (a *AudioService) Resample(data []byte, targetSampleRate uint32) ([]byte, error) {
	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read format: %w", err)
	}

	if format.SampleRate == targetSampleRate {
		return data, nil
	}

	var int16Samples []int16

	for {
		samples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read samples: %w", err)
		}

		for _, sample := range samples {
			sum := 0
			for ch := uint(0); ch < uint(format.NumChannels); ch++ {
				sum += wavReader.IntValue(sample, ch)
			}
			avgValue := int16(sum / int(format.NumChannels))
			int16Samples = append(int16Samples, avgValue)
		}
	}

	resampler, err := gomplerate.NewResampler(
		1,
		int(format.SampleRate),
		int(targetSampleRate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resampler: %w", err)
	}

	resampledSamples := resampler.ResampleInt16(int16Samples)

	outputSamples := make([]wav.Sample, len(resampledSamples))
	for i, val := range resampledSamples {
		outputSamples[i] = wav.Sample{
			Values: [2]int{int(val), 0},
		}
	}

	var outputBuffer bytes.Buffer
	writer := wav.NewWriter(
		&outputBuffer,
		uint32(len(outputSamples)),
		1,
		targetSampleRate,
		format.BitsPerSample,
	)

	if err := writer.WriteSamples(outputSamples); err != nil {
		return nil, fmt.Errorf("failed to write samples: %w", err)
	}

	return outputBuffer.Bytes(), nil
}

func findPeak(samples []int) int {
	if len(samples) == 0 {
		return 0
	}

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	peak := 0
	for _, sample := range samples {
		if abs(sample) > peak {
			peak = abs(sample)
		}
	}

	return peak
}

func (a *AudioService) Normalize(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read format: %w", err)
	}

	var samples []int
	for {
		wavSamples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read samples: %w", err)
		}

		for _, sample := range wavSamples {
			for ch := uint(0); ch < uint(format.NumChannels); ch++ {
				samples = append(samples, wavReader.IntValue(sample, ch))
			}
		}
	}

	peak := findPeak(samples)
	if peak == 0 {
		return data, nil // Silent audio
	}

	maxValue := int(1 << (format.BitsPerSample - 1))
	targetPeak := 0.95 * float64(maxValue)
	scaleFactor := targetPeak / float64(peak)

	if scaleFactor >= 0.99 && scaleFactor <= 1.01 {
		return data, nil
	}

	maxInt := int(maxValue) - 1
	minInt := -int(maxValue)
	normalizedSamples := make([]int, len(samples))

	for i, sample := range samples {
		scaled := float64(sample) * scaleFactor
		intValue := int(scaled)

		if intValue > maxInt {
			intValue = maxInt
		}
		if intValue < minInt {
			intValue = minInt
		}

		normalizedSamples[i] = intValue
	}

	numChannels := int(format.NumChannels)
	numSamples := len(normalizedSamples) / numChannels
	outputSamples := make([]wav.Sample, numSamples)

	for i := 0; i < numSamples; i++ {
		var values [2]int
		for ch := 0; ch < numChannels && ch < 2; ch++ {
			values[ch] = normalizedSamples[i*numChannels+ch]
		}
		outputSamples[i] = wav.Sample{Values: values}
	}

	var outputBuffer bytes.Buffer
	writer := wav.NewWriter(
		&outputBuffer,
		uint32(len(outputSamples)),
		format.NumChannels,
		format.SampleRate,
		format.BitsPerSample,
	)

	if err := writer.WriteSamples(outputSamples); err != nil {
		return nil, fmt.Errorf("failed to write samples: %w", err)
	}

	return outputBuffer.Bytes(), nil
}

func (a *AudioService) Spectrogram(data []byte, windowSize, hopSize int) (*Spectrogram, error) {
	reader := bytes.NewReader(data)
	wavReader := wav.NewReader(reader)
	format, err := wavReader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read format: %w", err)
	}

	var samples []float64
	maxValue := int(1 << (format.BitsPerSample - 1))

	for {
		wavSamples, err := wavReader.ReadSamples()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read samples: %w", err)
		}

		for _, sample := range wavSamples {
			sum := 0
			for ch := uint(0); ch < uint(format.NumChannels); ch++ {
				sum += wavReader.IntValue(sample, ch)
			}
			avg := float64(sum) / float64(format.NumChannels)
			samples = append(samples, avg/float64(maxValue))
		}
	}

	numFrames := (len(samples)-windowSize)/hopSize + 1
	spectrogram := make([][]float64, numFrames)
	timeFrames := make([]float64, numFrames)

	for frameIdx := 0; frameIdx < numFrames; frameIdx++ {
		start := frameIdx * hopSize
		end := start + windowSize
		if end > len(samples) {
			break
		}

		windowSamples := samples[start:end]

		hannWindow := make([]float64, windowSize)
		for i := 0; i < windowSize; i++ {
			hannWindow[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(windowSize-1)))
		}

		windowed := make([]float64, windowSize)
		for i := 0; i < windowSize; i++ {
			windowed[i] = windowSamples[i] * hannWindow[i]
		}

		complexWindow := make([]complex128, windowSize)
		for i, s := range windowed {
			complexWindow[i] = complex(s, 0)
		}

		spectrum := fft.FFT(complexWindow)

		magnitudes := make([]float64, windowSize/2)
		for i := 0; i < windowSize/2; i++ {
			real := real(spectrum[i])
			imag := imag(spectrum[i])
			magnitudes[i] = math.Sqrt(real*real + imag*imag)
		}

		spectrogram[frameIdx] = magnitudes
		timeFrames[frameIdx] = float64(start) / float64(format.SampleRate)
	}

	frequencyBins := make([]float64, windowSize/2)
	for i := 0; i < windowSize/2; i++ {
		frequencyBins[i] = float64(i) * float64(format.SampleRate) / float64(windowSize)
	}

	return &Spectrogram{
		Data:          spectrogram,
		FrequencyBins: frequencyBins,
		TimeFrames:    timeFrames,
		SampleRate:    format.SampleRate,
	}, nil
}
