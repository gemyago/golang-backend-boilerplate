package diag

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type OtelHTTPMiddleware func(http.Handler) http.Handler

type OtelMiddlewareFactoryDeps struct {
	dig.In

	metric.MeterProvider
	trace.TracerProvider
	Config
}

func NewOtelHTTPMiddleware(
	deps OtelMiddlewareFactoryDeps,
) OtelHTTPMiddleware { // coverage-ignore -- Little value in testing this factory function
	return func(next http.Handler) http.Handler {
		if !deps.Config.Enabled {
			return next
		}

		return otelhttp.NewHandler(
			next,

			// we will use route pattern or URI
			// but need to set something here
			"http-request",

			otelhttp.WithMeterProvider(deps.MeterProvider),
			otelhttp.WithTracerProvider(deps.TracerProvider),
			otelhttp.WithSpanNameFormatter(
				func(_ string, r *http.Request) string {
					if r.Pattern != "" {
						return r.Pattern
					}
					return r.RequestURI
				},
			),
		)
	}
}
