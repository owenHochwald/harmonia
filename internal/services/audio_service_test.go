package services

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestReadWAVProperties(t *testing.T) {

}
