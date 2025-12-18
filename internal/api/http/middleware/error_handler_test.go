package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestNewAppErrorHandler(t *testing.T) {
	handler := NewAppErrorHandler(telemetry.RootTestLogger())
	t.Run("NotFoundError sets 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err := app.NewErrNotFound("user", "123")
		handler(w, req, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidInputError sets 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err := app.NewErrInvalidInput("email", "invalid")
		handler(w, req, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ConflictError sets 409", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err := app.NewErrConflict("email", "exists")
		handler(w, req, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("other error sets 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err := errors.New("some error")
		handler(w, req, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
