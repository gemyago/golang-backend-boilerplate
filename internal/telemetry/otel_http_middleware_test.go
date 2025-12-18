package telemetry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestOtelHTTPMiddleware(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func() OtelMiddlewareFactoryDeps {
		return OtelMiddlewareFactoryDeps{
			TextMapPropagator: NewTextMapPropagator(),
			MeterProvider:     metricNoop.NewMeterProvider(),
			TracerProvider: sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(tracetest.NewInMemoryExporter()),
			),
			OTELConfig: OTELConfig{},
		}
	}

	t.Run("call next if not enabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.OTELConfig.Enabled = false

		middleware := NewOtelHTTPMiddleware(deps)

		handlerCalled := false
		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		})

		middlewareHandler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, fake.Internet().URL(), http.NoBody)
		recorder := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(recorder, req)
		assert.True(t, handlerCalled)
	})

	t.Run("call next if enabled and start span", func(t *testing.T) {
		deps := makeMockDeps()
		deps.OTELConfig.Enabled = true

		middleware := NewOtelHTTPMiddleware(deps)

		handlerCalled := false
		wantPattern := "GET " + fake.Lorem().Word()
		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			span := trace.SpanFromContext(req.Context())
			assert.True(t, span.SpanContext().IsValid())
			assert.Equal(t, wantPattern, req.Pattern)
			handlerCalled = true
		})

		middlewareHandler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, fake.Internet().URL(), http.NoBody)
		req.Pattern = wantPattern
		recorder := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(recorder, req)
		assert.True(t, handlerCalled)

		assert.NotEmpty(t, recorder.Header().Get("traceparent"))
	})
}
