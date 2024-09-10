package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckRoutes(t *testing.T) {
	type mockDeps struct {
		Deps
		Mux *http.ServeMux
	}
	makeDeps := func() mockDeps {
		mux := http.NewServeMux()
		deps := Deps{
			RootLogger: diag.RootTestLogger(),
		}
		return mockDeps{
			Deps: deps,
			Mux:  mux,
		}
	}

	t.Run("GET /health", func(t *testing.T) {
		t.Run("should respond with OK", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
			w := httptest.NewRecorder()
			deps := makeDeps()
			MountHealthCheckRoutes(deps.Mux, deps.Deps)
			deps.Mux.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "OK", w.Body.String())
		})
	})
}
