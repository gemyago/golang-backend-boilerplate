package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationMiddleware(t *testing.T) {
	fake := faker.New()
	t.Run("set new correlation id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		res := httptest.NewRecorder()
		idGen := ident.NewMockGenerator()
		mw := NewCorrelationMiddleware(idGen)
		nextCalled := false
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logAttributes := diag.GetLogAttributesFromContext(r.Context())
			wantCorrelationID := ident.MockGeneratorLastGenerated(idGen)
			assert.Equal(t, wantCorrelationID.String(), logAttributes.CorrelationID.String())
			assert.Equal(t, wantCorrelationID.String(), w.Header().Get(diag.CorrelationIDHeader))
			nextCalled = true
		})).ServeHTTP(res, req)
		assert.True(t, nextCalled)
	})
	t.Run("use existing correlation id", func(t *testing.T) {
		wantCorrelationID := fake.UUID().V4()
		req := httptest.NewRequest(http.MethodGet, "/something", http.NoBody)
		req.Header.Add(diag.CorrelationIDHeader, wantCorrelationID)
		res := httptest.NewRecorder()
		mw := NewCorrelationMiddleware(ident.NewDefaultGenerator())
		nextCalled := false
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logAttributes := diag.GetLogAttributesFromContext(r.Context())
			assert.Equal(t, wantCorrelationID, logAttributes.CorrelationID.String())
			assert.Equal(t, wantCorrelationID, w.Header().Get(diag.CorrelationIDHeader))
			nextCalled = true
		})).ServeHTTP(res, req)
		assert.True(t, nextCalled)
	})
}
