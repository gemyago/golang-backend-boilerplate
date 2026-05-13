package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddlewareDeps contains dependencies for the logging middleware.
type LoggingMiddlewareDeps struct {
	RootLogger *slog.Logger
}

// LoggingMiddleware wraps an [http.RoundTripper] to add structured logging.
type LoggingMiddleware struct {
	transport                 http.RoundTripper
	logger                    *slog.Logger
	obfuscatedRequestHeaders  map[string]struct{}
	obfuscatedResponseHeaders map[string]struct{}
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(transport http.RoundTripper, deps LoggingMiddlewareDeps) http.RoundTripper {
	return &LoggingMiddleware{
		transport: transport,
		logger:    deps.RootLogger.WithGroup("http-logging-middleware"),

		// Request headers to obfuscate in logs. Extend as needed.
		obfuscatedRequestHeaders: map[string]struct{}{
			"authorization": {},
			"cookie":        {},
			"set-cookie":    {},
			"x-auth-token":  {},
			"x-csrf-token":  {},
			"x-xsrf-token":  {},
		},

		// Response headers to obfuscate in logs. Extend as needed.
		obfuscatedResponseHeaders: map[string]struct{}{
			"set-cookie": {},
		},
	}
}

// RoundTrip implements the [http.RoundTripper] interface.
// Logs request and response details with structured logging.
func (l *LoggingMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Call next transport
	resp, err := l.transport.RoundTrip(req)
	duration := time.Since(start)

	requestHeadersGroup := buildObfuscatedHeadersAttr(req.Header, l.obfuscatedRequestHeaders)

	requestAttr := slog.Group("request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		requestHeadersGroup,
	)

	if err != nil {
		attrs := []slog.Attr{
			requestAttr,
			slog.Group("response", slog.Duration("duration", duration)),
			slog.Any("error", err),
		}

		// We still do it with warn level. Upper most layer should log with error
		l.logger.LogAttrs(req.Context(), slog.LevelWarn, "OUTBOUND_REQUEST_FAILED",
			attrs...,
		)
		return nil, err
	}

	responseHeadersGroup := buildObfuscatedHeadersAttr(resp.Header, l.obfuscatedResponseHeaders)

	level := slog.LevelDebug

	// We log everything above 400 as warnings for better visibility
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		level = slog.LevelWarn
	}

	attrs := []slog.Attr{
		requestAttr,
		slog.Group("response",
			slog.Int("status", resp.StatusCode),
			slog.Duration("duration", duration),
			responseHeadersGroup,
		),
	}

	l.logger.LogAttrs(req.Context(), level, "OUTBOUND_REQUEST_COMPLETED",
		attrs...,
	)

	return resp, nil
}

func buildObfuscatedHeadersAttr(headers http.Header, obfuscatedHeaders map[string]struct{}) slog.Attr {
	headerAttrs := make([]slog.Attr, 0, len(headers))
	for key, values := range headers {
		var headerAttr slog.Attr
		if _, ok := obfuscatedHeaders[strings.ToLower(key)]; ok {
			headerAttr = slog.Any(key, []string{"[REDACTED]"})
		} else {
			headerAttr = slog.Any(key, values)
		}
		headerAttrs = append(headerAttrs, headerAttr)
	}
	return slog.GroupAttrs("headers", headerAttrs...)
}
