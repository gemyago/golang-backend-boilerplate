package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

// CorrelationMiddlewareDeps contains dependencies for the correlation middleware.
type CorrelationMiddlewareDeps struct{}

// CorrelationMiddleware wraps an http.RoundTripper to add correlation ID to outbound requests.
type CorrelationMiddleware struct {
	transport http.RoundTripper
}

// NewCorrelationMiddleware creates a new correlation middleware.
func NewCorrelationMiddleware(transport http.RoundTripper) http.RoundTripper {
	return &CorrelationMiddleware{
		transport: transport,
	}
}

// RoundTrip implements http.RoundTripper interface.
// Adds correlation ID header to the request.
func (c *CorrelationMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	logAttrs := diag.GetLogAttributesFromContext(req.Context())
	if logAttrs.CorrelationID.Kind() == slog.KindString {
		if correlationID := logAttrs.CorrelationID.String(); correlationID != "" {
			req.Header.Set(diag.CorrelationIDHeader, correlationID)
		}
	}
	return c.transport.RoundTrip(req)
}
