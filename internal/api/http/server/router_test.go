package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaswdr/faker"
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

		adapter := NewHTTPRouter(HTTPRouterDeps{})
		handlerInvoked := false
		adapter.HandleRoute(
			http.MethodGet,
			"/resources/{param}/value",
			http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotPathParam := adapter.PathValue(r, "param")
				assert.Equal(t, wantPathParam, gotPathParam)
				handlerInvoked = true
			}))
		adapter.ServeHTTP(httptest.NewRecorder(), req)
		assert.True(t, handlerInvoked)
	})
}
