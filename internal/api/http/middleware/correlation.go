package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
)

// NewCorrelationMiddleware creates a middleware that sets a correlation ID in the request context.
// Otel may not always be enabled and we want this additional mechanism to be always in place.
func NewCorrelationMiddleware(idGen ident.Generator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			correlationID := req.Header.Get(telemetry.CorrelationIDHeader)
			if correlationID == "" {
				correlationID = idGen.MustNewV7().String()
			}
			logAttributes := telemetry.GetLogAttributesFromContext(req.Context())
			logAttributes.CorrelationID = slog.StringValue(correlationID)
			nextCtx := telemetry.SetLogAttributesToContext(req.Context(), logAttributes)

			w.Header().Set(telemetry.CorrelationIDHeader, correlationID)

			next.ServeHTTP(w, req.WithContext(nextCtx))
		})
	}
}
