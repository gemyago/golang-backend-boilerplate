package v1controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	newHandler := func() http.Handler {
		return server.NewTestRootHandler().
			RegisterHealthRoutes(HealthController{})
	}

	t.Run("GET /health", func(t *testing.T) {
		t.Run("should respond with OK", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			w := httptest.NewRecorder()
			newHandler().ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, `{"status": "OK"}`, w.Body.String())
		})
	})
}
