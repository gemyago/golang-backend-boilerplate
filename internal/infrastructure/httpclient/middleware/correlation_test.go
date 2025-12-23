package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelationMiddleware_RoundTrip(t *testing.T) {
	faker := faker.New()

	t.Run("should set correlation ID header from context", func(t *testing.T) {
		// Given
		expectedCorrelationID := faker.UUID().V4()
		ctx := context.Background()
		logAttrs := telemetry.GetLogAttributesFromContext(ctx)
		logAttrs.CorrelationID = slog.StringValue(expectedCorrelationID)
		ctx = telemetry.SetLogAttributesToContext(ctx, logAttrs)

		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil).WithContext(ctx)

		mockTransport := &mockRoundTripper{}
		middleware := NewCorrelationMiddleware(mockTransport)

		// When
		_, err := middleware.RoundTrip(req)

		// Then
		require.NoError(t, err)
		assert.Equal(t, expectedCorrelationID, req.Header.Get(telemetry.CorrelationIDHeader))
	})

	t.Run("should overwrite existing correlation ID header", func(t *testing.T) {
		// Given
		expectedCorrelationID := faker.UUID().V4()
		existingHeaderValue := faker.UUID().V4()
		ctx := context.Background()
		logAttrs := telemetry.GetLogAttributesFromContext(ctx)
		logAttrs.CorrelationID = slog.StringValue(expectedCorrelationID)
		ctx = telemetry.SetLogAttributesToContext(ctx, logAttrs)

		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil).WithContext(ctx)
		req.Header.Set(telemetry.CorrelationIDHeader, existingHeaderValue)

		mockTransport := &mockRoundTripper{}
		middleware := NewCorrelationMiddleware(mockTransport)

		// When
		_, err := middleware.RoundTrip(req)

		// Then
		require.NoError(t, err)
		assert.Equal(t, expectedCorrelationID, req.Header.Get(telemetry.CorrelationIDHeader))
	})

	t.Run("should not set header when no correlation ID in context", func(t *testing.T) {
		// Given
		ctx := context.Background()
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil).WithContext(ctx)

		mockTransport := &mockRoundTripper{}
		middleware := NewCorrelationMiddleware(mockTransport)

		// When
		_, err := middleware.RoundTrip(req)

		// Then
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get(telemetry.CorrelationIDHeader))
	})
}

type mockRoundTripper struct{}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}
