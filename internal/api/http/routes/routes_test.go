package routes

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

type errWriter func(data []byte) (int, error)

func (w errWriter) Write(data []byte) (int, error) {
	return w(data)
}

func TestWriteData(t *testing.T) {
	t.Run("should write", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		log := diag.RootTestLogger()
		wantData := faker.Sentence()
		var writer bytes.Buffer
		WriteData(req, log, &writer, []byte(wantData))
		assert.Equal(t, wantData, writer.String())
	})
	t.Run("should log error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		log := diag.RootTestLogger()
		wantData := faker.Sentence()
		wantErr := errors.New(faker.Sentence())

		assert.NotPanics(t, func() {
			WriteData(req, log, errWriter(func(_ []byte) (int, error) {
				return 0, wantErr
			}), []byte(wantData))
		})
	})
}
