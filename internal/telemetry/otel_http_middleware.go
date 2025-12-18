package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type OtelHTTPMiddleware func(http.Handler) http.Handler

type OtelMiddlewareFactoryDeps struct {
	dig.In

	metric.MeterProvider
	trace.TracerProvider
	propagation.TextMapPropagator
	OTELConfig
}

// injectOtelResponseHeaders injects OTel trace context headers (traceparent) to the HTTP response.
// This sometimes helps with debugging and tracing, especially with Postman, or directly from the browser.
func injectOtelResponseHeaders(
	deps OtelMiddlewareFactoryDeps,
	next http.Handler,
) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		spanCtx := trace.SpanContextFromContext(ctx)

		if spanCtx.IsValid() {
			deps.TextMapPropagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))
		}

		next.ServeHTTP(w, r)
	})
}

func NewOtelHTTPMiddleware(
	deps OtelMiddlewareFactoryDeps,
) OtelHTTPMiddleware { // coverage-ignore -- Little value in testing this factory function
	return func(next http.Handler) http.Handler {
		if !deps.OTELConfig.Enabled {
			return next
		}

		resultingHandler := otelhttp.NewHandler(
			injectOtelResponseHeaders(deps, next),

			// span name will be set by the span name formatter below
			// we will use route pattern or URI
			// but need to set something here
			"http-request",

			otelhttp.WithPropagators(deps.TextMapPropagator),
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

		return resultingHandler
	}
}
