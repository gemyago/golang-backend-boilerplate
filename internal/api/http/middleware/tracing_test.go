package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestTracingMiddleware(t *testing.T) {
	fake := faker.New()
	t.Run("set new correlation id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		res := httptest.NewRecorder()
		mw := NewTracingMiddleware(NewTracingMiddlewareCfg())
		nextCalled := false
		mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			logAttributes := diag.GetLogAttributesFromContext(r.Context())
			assert.NotEmpty(t, logAttributes.CorrelationID.String())
			nextCalled = true
		})).ServeHTTP(res, req)
		assert.True(t, nextCalled)
	})
	t.Run("use existing correlation id", func(t *testing.T) {
		wantCorrelationID := fake.UUID().V4()
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		req.Header.Add("X-Correlation-ID", wantCorrelationID)
		res := httptest.NewRecorder()
		mw := NewTracingMiddleware(NewTracingMiddlewareCfg())
		nextCalled := false
		mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			logAttributes := diag.GetLogAttributesFromContext(r.Context())
			assert.Equal(t, wantCorrelationID, logAttributes.CorrelationID.String())
			nextCalled = true
		})).ServeHTTP(res, req)
		assert.True(t, nextCalled)
	})
}
