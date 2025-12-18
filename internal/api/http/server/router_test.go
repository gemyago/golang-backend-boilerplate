package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

func TestMuxRouterAdapter(t *testing.T) {
	fake := faker.New()
	t.Run("should handle routes and read path values", func(t *testing.T) {
		wantPathParam := fake.Lorem().Word()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/resources/%s/value", wantPathParam),
			http.NoBody,
		)

		calls := []string{}
		adapter := NewHTTPRouter(HTTPRouterDeps{
			Middleware: func(h http.Handler) http.Handler {
				calls = append(calls, "middleware")
				return h
			},
		})
		adapter.HandleRoute(
			http.MethodGet,
			"/resources/{param}/value",
			http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotPathParam := adapter.PathValue(r, "param")
				assert.Equal(t, wantPathParam, gotPathParam)
				calls = append(calls, "handler")
			}))
		adapter.ServeHTTP(httptest.NewRecorder(), req)
		assert.Equal(t, []string{"middleware", "handler"}, calls)
	})
}
