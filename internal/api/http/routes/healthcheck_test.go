package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckRoutes(t *testing.T) {
	type mockDeps struct {
		HealthCheckDeps
		Mux *http.ServeMux
	}
	makeDeps := func() mockDeps {
		mux := http.NewServeMux()
		deps := HealthCheckDeps{
			RootLogger: diag.RootTestLogger(),
		}
		return mockDeps{
			HealthCheckDeps: deps,
			Mux:             mux,
		}
	}

	t.Run("GET /health", func(t *testing.T) {
		t.Run("should respond with OK", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			w := httptest.NewRecorder()
			deps := makeDeps()
			NewHealthCheckRoutesGroup(deps.HealthCheckDeps).Mount(deps.Mux)
			deps.Mux.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "OK", w.Body.String())
		})
	})
}
