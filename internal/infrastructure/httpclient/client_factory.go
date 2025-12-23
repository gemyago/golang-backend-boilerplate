package httpclient

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient/middleware"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"go.uber.org/dig"
	"golang.org/x/oauth2"
)

const (
	// defaultClientTimeout is the default timeout for HTTP clients.
	defaultClientTimeout = 30 * time.Second

	defaultMaxIdleConns          = 100
	defaultIdleConnTimeout       = 90 * time.Second
	defaultTLSHandshakeTimeout   = 10 * time.Second
	defaultExpectContinueTimeout = 1 * time.Second
)

// ClientFactoryDeps contains dependencies for the client factory.
type ClientFactoryDeps struct {
	dig.In

	RootLogger *slog.Logger

	OtelHTTPTransportFactory telemetry.OtelHTTPTransportFactory
}

// ClientOption configures HTTP client creation.
type ClientOption func(*clientConfig)

// clientConfig holds internal configuration for HTTP client creation.
type clientConfig struct {
	timeout         time.Duration
	authTokenSource oauth2.TokenSource
	enableLogging   bool
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithAuthTokenSource sets the OAuth2 token source for authentication.
// Usually set to oauth2.StaticTokenSource for basic scenarios with fixed tokens.
func WithAuthTokenSource(tokenSource oauth2.TokenSource) ClientOption {
	return func(c *clientConfig) {
		c.authTokenSource = tokenSource
	}
}

// WithLogging sets whether logging middleware is enabled.
func WithLogging(enabled bool) ClientOption {
	return func(c *clientConfig) {
		c.enableLogging = enabled
	}
}

// ClientFactory is responsible for creating configured HTTP clients with middleware.
type ClientFactory struct {
	logger                   *slog.Logger
	otelHTTPTransportFactory telemetry.OtelHTTPTransportFactory
}

// NewClientFactory creates a new client factory.
func NewClientFactory(deps ClientFactoryDeps) *ClientFactory {
	otelHTTPFactory := deps.OtelHTTPTransportFactory
	if otelHTTPFactory == nil {
		otelHTTPFactory = func(next http.RoundTripper) http.RoundTripper {
			return next
		}
	}
	return &ClientFactory{
		logger:                   deps.RootLogger.WithGroup("http-client-factory"),
		otelHTTPTransportFactory: otelHTTPFactory,
	}
}

// CreateClient creates a new HTTP client with the specified options.
// Middleware is applied in the order: Correlation -> Logging -> Auth -> Otel -> BaseTransport
// This ensures correlation ID is set first, then logging captures the full request lifecycle, auth adds headers, and otel traces.
func (f *ClientFactory) CreateClient(options ...ClientOption) *http.Client {
	config := &clientConfig{
		timeout:       defaultClientTimeout,
		enableLogging: true, // Default: enabled
	}

	for _, option := range options {
		option(config)
	}

	// Start with the base transport
	var transport http.RoundTripper = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          defaultMaxIdleConns,
		IdleConnTimeout:       defaultIdleConnTimeout,
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
		ExpectContinueTimeout: defaultExpectContinueTimeout,
	}

	// Middleware below applied in a reverse order of execution

	// Logging middleware is outermost to capture full request lifecycle
	if config.enableLogging {
		transport = middleware.NewLoggingMiddleware(transport, middleware.LoggingMiddlewareDeps{
			RootLogger: f.logger,
		})
	}

	if config.authTokenSource != nil {
		transport = &oauth2.Transport{
			Source: config.authTokenSource,
			Base:   transport,
		}
	}

	// We still want to keep correlation just in case otel is not enabled/available
	transport = middleware.NewCorrelationMiddleware(transport)

	// Enabling/disabling it is controlled globally (in config)
	// Add option if you need per client control
	transport = f.otelHTTPTransportFactory(transport)

	return &http.Client{
		Transport: transport,
		Timeout:   config.timeout,
	}
}
