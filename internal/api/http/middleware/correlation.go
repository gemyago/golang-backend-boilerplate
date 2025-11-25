package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/gofrs/uuid/v5"
)

const CorrelationIDHeader = "X-Correlation-ID"

type CorrelationMiddlewareCfg struct {
	generateUUID func() string
}

func NewCorrelationMiddlewareCfg() *CorrelationMiddlewareCfg {
	return &CorrelationMiddlewareCfg{
		generateUUID: func() string {
			return uuid.Must(uuid.NewV4()).String()
		},
	}
}

// NewCorrelationMiddleware creates a middleware that sets a correlation ID in the request context.
// Otel may not always be enabled and we want this additional mechanism to be always in place.
func NewCorrelationMiddleware(cfg *CorrelationMiddlewareCfg) Middleware {
	generateUUID := cfg.generateUUID
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			correlationID := req.Header.Get(CorrelationIDHeader)
			if correlationID == "" {
				correlationID = generateUUID()
			}
			logAttributes := diag.GetLogAttributesFromContext(req.Context())
			logAttributes.CorrelationID = slog.StringValue(correlationID)
			nextCtx := diag.SetLogAttributesToContext(req.Context(), logAttributes)

			w.Header().Set(CorrelationIDHeader, correlationID)

			next.ServeHTTP(w, req.WithContext(nextCtx))
		})
	}
}
