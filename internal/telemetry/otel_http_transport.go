package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type OtelHTTPTransportFactory = func(http.RoundTripper) http.RoundTripper

type OtelHTTPTransportFactoryDeps struct {
	dig.In

	metric.MeterProvider
	trace.TracerProvider
	propagation.TextMapPropagator
	OTELConfig
}

func NewOtelHTTPTransportFactory(
	deps OtelHTTPTransportFactoryDeps,
) OtelHTTPTransportFactory { // coverage-ignore -- Little value in testing wire-up mostly code
	return func(next http.RoundTripper) http.RoundTripper {
		if !deps.OTELConfig.Enabled {
			return next
		}

		return otelhttp.NewTransport(
			next,
			otelhttp.WithPropagators(deps.TextMapPropagator),
			otelhttp.WithMeterProvider(deps.MeterProvider),
			otelhttp.WithTracerProvider(deps.TracerProvider),

			// Default is just "HTTP {method}" which is not very useful
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return "HTTP " + r.Method + " " + r.URL.String()
			}),
		)
	}
}
