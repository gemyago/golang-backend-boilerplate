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

	// Log response
	if err != nil {
		l.logger.ErrorContext(req.Context(), "HTTP request failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Duration("duration", duration),
			slog.Any("error", err),
		)
		return nil, err
	}

	requestHeaders := make([]slog.Attr, 0, len(req.Header))
	for key, values := range req.Header {
		requestHeaders = append(requestHeaders, slog.Any(key, values))
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
		slog.Group("request",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.GroupAttrs("headers", requestHeaders...),
		),
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
