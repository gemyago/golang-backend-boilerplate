package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddlewareDeps contains dependencies for the logging middleware.
type LoggingMiddlewareDeps struct {
	RootLogger *slog.Logger
}

// LoggingMiddleware wraps an http.RoundTripper to add structured logging.
type LoggingMiddleware struct {
	transport http.RoundTripper
	logger    *slog.Logger
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(transport http.RoundTripper, deps LoggingMiddlewareDeps) http.RoundTripper {
	return &LoggingMiddleware{
		transport: transport,
		logger:    deps.RootLogger.WithGroup("http-logging-middleware"),
	}
}

// RoundTrip implements http.RoundTripper interface.
// Logs request and response details with structured logging.
func (l *LoggingMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Call next transport
	resp, err := l.transport.RoundTrip(req)
	duration := time.Since(start)

	requestHeaders := make([]slog.Attr, 0, len(req.Header))
	for key, values := range req.Header {
		requestHeaders = append(requestHeaders, slog.Any(key, values))
	}

	requestAttr := slog.Group("request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.GroupAttrs("headers", requestHeaders...),
	)

	// Log response
	if err != nil {
		attrs := []slog.Attr{
			requestAttr,
			slog.Group("response", slog.Duration("duration", duration)),
			slog.Any("error", err),
		}

		// We still do it with warn level. Upper most layer should log with error
		l.logger.LogAttrs(req.Context(), slog.LevelWarn, "OUTBOUND_CALL_FAILED",
			attrs...,
		)
		return nil, err
	}

	responseHeaders := make([]slog.Attr, 0, len(resp.Header))
	for key, values := range resp.Header {
		responseHeaders = append(responseHeaders, slog.Any(key, values))
	}

	level := slog.LevelDebug

	// We log everything above 400 as warnings for better visibility
	if resp.StatusCode >= 400 && resp.StatusCode < 599 {
		level = slog.LevelWarn
	}

	attrs := []slog.Attr{
		requestAttr,
		slog.Group("response",
			slog.Int("status", resp.StatusCode),
			slog.Duration("duration", duration),
			slog.GroupAttrs("headers", responseHeaders...),
		),
	}

	l.logger.LogAttrs(req.Context(), level, "OUTBOUND_CALL_COMPLETED",
		attrs...,
	)

	return resp, nil
}
