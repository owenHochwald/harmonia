package services

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/youpy/go-wav"
)

func TestValidateFile(t *testing.T) {
	service := AudioService{}

	t.Run("valid WAV file", func(t *testing.T) {
		// TODO: join with curr dir for relative path rather than an absolute path
		audioFile := "/Users/owenhochwald/Documents/code/personal/backend/go/harmonia/public/audios/sample-12s.wav"
		file, err := os.Open(audioFile)
		if err != nil {
			t.Skipf("Audio file not found - skipping test")
		}

		data, _ := io.ReadAll(file)
		reader := bytes.NewReader(data)

		err, code := service.ValidateFile(reader)

		assert.Nil(t, err)
		assert.Equal(t, 200, code)
	})

	t.Run("wrong file type", func(t *testing.T) {
		audioFile := "/Users/owenhochwald/Documents/code/personal/backend/go/harmonia/README.md"
		file, err := os.Open(audioFile)
		if err != nil {
			t.Skipf("Audio file not found - skipping test")
		}

		data, _ := io.ReadAll(file)
		reader := bytes.NewReader(data)

		err, code := service.ValidateFile(reader)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "error reading WAV format")
		assert.Equal(t, http.StatusInternalServerError, code)
	})

	t.Run("file too large", func(t *testing.T) {
		data := make([]byte, 11*1024*1024)
		reader := bytes.NewReader(data)

		err, code := service.ValidateFile(reader)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "file is too large")
		assert.Equal(t, http.StatusBadRequest, code)
	})
}

func createTestWAV(t *testing.T, sampleRate uint32, channels uint16, samples []wav.Sample) []byte {
	var buf bytes.Buffer
	writer := wav.NewWriter(&buf, uint32(len(samples)), channels, sampleRate, 16)

	err := writer.WriteSamples(samples)
	require.NoError(t, err)

	require.NoError(t, err)

	return buf.Bytes()
}

func TestConvertToMono(t *testing.T) {
	service := NewAudioService()

	t.Run("Already mono returns same data", func(t *testing.T) {
		// Create mono WAV
		samples := []wav.Sample{
			{Values: [2]int{1000, 0}},
			{Values: [2]int{-500, 0}},
			{Values: [2]int{2000, 0}},
		}
		monoWAV := createTestWAV(t, 44100, 1, samples)

		result, err := service.ConvertToMono(monoWAV)
		require.NoError(t, err)
		assert.Equal(t, monoWAV, result)
	})

	t.Run("Stereo converts to mono", func(t *testing.T) {
		// Create stereo WAV
		samples := []wav.Sample{
			{Values: [2]int{1000, 2000}},  // Avg = 1500
			{Values: [2]int{-500, -1500}}, // Avg = -1000
		}
		stereoWAV := createTestWAV(t, 44100, 2, samples)

		result, err := service.ConvertToMono(stereoWAV)
		require.NoError(t, err)

		reader := bytes.NewReader(result)
		wavReader := wav.NewReader(reader)
		format, _ := wavReader.Format()
		assert.Equal(t, uint16(1), format.NumChannels)
	})
}

func TestResample(t *testing.T) {
	service := NewAudioService()

	t.Run("Same rate returns original", func(t *testing.T) {
		samples := make([]wav.Sample, 44100) // 1 second at 44.1kHz
		for i := range samples {
			samples[i] = wav.Sample{Values: [2]int{int(i % 1000), 0}}
		}
		wavData := createTestWAV(t, 44100, 1, samples)

		result, err := service.Resample(wavData, 44100)
		require.NoError(t, err)
		assert.Equal(t, wavData, result)
	})

	t.Run("Downsample 44100 to 16000", func(t *testing.T) {
		samples := make([]wav.Sample, 44100) // 1 second
		for i := range samples {
			samples[i] = wav.Sample{Values: [2]int{1000, 0}}
		}
		wavData := createTestWAV(t, 44100, 1, samples)

		result, err := service.Resample(wavData, 16000)
		require.NoError(t, err)

		reader := bytes.NewReader(result)
		wavReader := wav.NewReader(reader)
		format, _ := wavReader.Format()
		assert.Equal(t, uint32(16000), format.SampleRate)

		totalSamples, _ := service.GetTotalSamples(result)
		assert.InDelta(t, 16000, totalSamples, 800)
	})
}

func TestNormalize(t *testing.T) {
	service := NewAudioService()

	t.Run("Silent audio returns unchanged", func(t *testing.T) {
		samples := []wav.Sample{
			{Values: [2]int{0, 0}},
			{Values: [2]int{0, 0}},
		}
		wavData := createTestWAV(t, 44100, 1, samples)

		result, err := service.Normalize(wavData)
		require.NoError(t, err)
		assert.Equal(t, wavData, result)
	})

	t.Run("Quiet audio gets boosted", func(t *testing.T) {
		samples := []wav.Sample{
			{Values: [2]int{500, 0}},
			{Values: [2]int{1000, 0}}, // Peak
			{Values: [2]int{-800, 0}},
		}
		wavData := createTestWAV(t, 44100, 1, samples)

		result, err := service.Normalize(wavData)
		require.NoError(t, err)

		reader := bytes.NewReader(result)
		wavReader := wav.NewReader(reader)

		var normalizedSamples []int
		for {
			samples, err := wavReader.ReadSamples()
			if err != nil {
				break
			}
			for _, s := range samples {
				normalizedSamples = append(normalizedSamples, wavReader.IntValue(s, 0))
			}
		}

		newPeak := findPeak(normalizedSamples)

		assert.InDelta(t, 31129, newPeak, 100)
	})

	t.Run("Loud audio gets reduced", func(t *testing.T) {
		samples := []wav.Sample{
			{Values: [2]int{32000, 0}}, // Peak
			{Values: [2]int{-20000, 0}},
		}
		wavData := createTestWAV(t, 44100, 1, samples)

		result, err := service.Normalize(wavData)
		require.NoError(t, err)

		reader := bytes.NewReader(result)
		wavReader := wav.NewReader(reader)

		var normalizedSamples []int
		for {
			samples, err := wavReader.ReadSamples()
			if err != nil {
				break
			}
			for _, s := range samples {
				normalizedSamples = append(normalizedSamples, wavReader.IntValue(s, 0))
			}
		}

		newPeak := findPeak(normalizedSamples)
		assert.InDelta(t, 31129, newPeak, 100)
	})
}

func TestSpectrogram(t *testing.T) {
	service := NewAudioService()

	t.Run("Generates spectrogram with correct dimensions", func(t *testing.T) {
		sampleRate := 16000
		duration := 1.0
		frequency := 440.0
		numSamples := int(float64(sampleRate) * duration)

		samples := make([]wav.Sample, numSamples)
		for i := 0; i < numSamples; i++ {
			t := float64(i) / float64(sampleRate)
			value := int(10000 * math.Sin(2*math.Pi*frequency*t))
			samples[i] = wav.Sample{Values: [2]int{value, 0}}
		}

		wavData := createTestWAV(t, uint32(sampleRate), 1, samples)

		windowSize := 2048
		hopSize := 512
		spec, err := service.Spectrogram(wavData, windowSize, hopSize)
		require.NoError(t, err)

		expectedFrames := (numSamples-windowSize)/hopSize + 1
		assert.Equal(t, expectedFrames, len(spec.Data), "Should have correct number of frames")
		assert.Equal(t, windowSize/2, len(spec.FrequencyBins), "Should have windowSize/2 frequency bins")
		assert.Equal(t, len(spec.Data), len(spec.TimeFrames), "Time frames should match data frames")

		for i, frame := range spec.Data {
			assert.Equal(t, windowSize/2, len(frame), "Frame %d should have correct size", i)
		}

		expectedFreqResolution := float64(sampleRate) / float64(windowSize)
		for i, freq := range spec.FrequencyBins {
			expectedFreq := float64(i) * expectedFreqResolution
			assert.InDelta(t, expectedFreq, freq, 0.1, "Frequency bin %d should be correct", i)
		}

		expectedTimeStep := float64(hopSize) / float64(sampleRate)
		for i := 1; i < len(spec.TimeFrames); i++ {
			timeDiff := spec.TimeFrames[i] - spec.TimeFrames[i-1]
			assert.InDelta(t, expectedTimeStep, timeDiff, 0.001, "Time step should be consistent")
		}

		t.Logf("Spectrogram: %d frames × %d freq bins", len(spec.Data), len(spec.FrequencyBins))
		t.Logf("Frequency resolution: %.2f Hz", expectedFreqResolution)
		t.Logf("Time resolution: %.4f seconds", expectedTimeStep)
	})

	t.Run("Detects frequency peak", func(t *testing.T) {
		sampleRate := 16000
		frequency := 1000.0
		numSamples := 16000 // 1 second

		samples := make([]wav.Sample, numSamples)
		for i := 0; i < numSamples; i++ {
			t := float64(i) / float64(sampleRate)
			value := int(15000 * math.Sin(2*math.Pi*frequency*t))
			samples[i] = wav.Sample{Values: [2]int{value, 0}}
		}

		wavData := createTestWAV(t, uint32(sampleRate), 1, samples)

		spec, err := service.Spectrogram(wavData, 2048, 512)
		require.NoError(t, err)

		freqResolution := float64(sampleRate) / 2048.0
		expectedBin := int(frequency / freqResolution)

		middleFrame := spec.Data[len(spec.Data)/2]

		peakBin := 0
		peakMagnitude := 0.0
		for i, mag := range middleFrame {
			if mag > peakMagnitude {
				peakMagnitude = mag
				peakBin = i
			}
		}

		assert.InDelta(t, expectedBin, peakBin, 2, "Peak should be near 1000Hz")

		peakFreq := spec.FrequencyBins[peakBin]
		t.Logf("Expected peak: %.0f Hz (bin %d)", frequency, expectedBin)
		t.Logf("Detected peak: %.0f Hz (bin %d)", peakFreq, peakBin)
		t.Logf("Peak magnitude: %.2f", peakMagnitude)
	})
}

func TestProcess(t *testing.T) {
	service := NewAudioService()

	t.Run("Full pipeline", func(t *testing.T) {
		samples := make([]wav.Sample, 44100) // 1 second
		for i := range samples {
			samples[i] = wav.Sample{
				Values: [2]int{
					int(5000 * float64(i) / 44100),
					int(3000 * float64(i) / 44100),
				},
			}
		}
		wavData := createTestWAV(t, 44100, 2, samples)

		result, err := service.Process(wavData)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, uint16(2), result.Metadata.OriginalChannels)
		assert.Equal(t, uint32(44100), result.Metadata.OriginalRate)
	})
}

func TestWithRealFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real file test")
	}

	data, err := os.ReadFile("testdata/test.wav")
	if err != nil {
		t.Skip("No test file found")
	}

	service := NewAudioService()
	result, err := service.Process(data)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Processed: %+v", result.Metadata)
}
