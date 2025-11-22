package otel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type HTTPMiddlewareFactory func(
	operation string,
) func(http.Handler) http.Handler

func NewNoopOTELMiddlewareFactory() HTTPMiddlewareFactory {
	return func(string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
}

type MiddlewareFactoryDeps struct {
	dig.In

	metric.MeterProvider
	trace.TracerProvider
	Config
}

func NewOTELMiddlewareFactory(
	deps MiddlewareFactoryDeps,
) HTTPMiddlewareFactory {
	return func(operation string) func(http.Handler) http.Handler {
		if !deps.Config.Enabled {
			return NewNoopOTELMiddlewareFactory()(operation)
		}

		return otelhttp.NewMiddleware(
			operation,
			otelhttp.WithMeterProvider(deps.MeterProvider),
			otelhttp.WithTracerProvider(deps.TracerProvider),
		)
	}
}
