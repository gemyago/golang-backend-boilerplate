package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
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
}

// ClientOption configures HTTP client creation.
type ClientOption func(*clientConfig)

// clientConfig holds internal configuration for HTTP client creation.
type clientConfig struct {
	timeout             time.Duration
	authTokenSource     oauth2.TokenSource
	enableLogging       bool
	enableErrorHandling bool
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

// WithErrorHandling sets whether error handling middleware is enabled.
func WithErrorHandling(enabled bool) ClientOption {
	return func(c *clientConfig) {
		c.enableErrorHandling = enabled
	}
}

// ClientFactory is responsible for creating configured HTTP clients with middleware.
type ClientFactory struct {
	logger *slog.Logger
}

// NewClientFactory creates a new client factory.
func NewClientFactory(deps ClientFactoryDeps) *ClientFactory {
	return &ClientFactory{
		logger: deps.RootLogger.WithGroup("http-client-factory"),
	}
}

// CreateClient creates a new HTTP client with the specified options.
// Middleware is applied in the order: Logging -> Auth -> ErrorHandling -> BaseTransport
// This ensures logging captures the full request lifecycle, auth adds headers, and error handling catches issues.
func (f *ClientFactory) CreateClient(options ...ClientOption) *http.Client {
	config := &clientConfig{
		timeout:             defaultClientTimeout,
		enableLogging:       true, // Default: enabled
		enableErrorHandling: true, // Default: enabled
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

	// Apply middleware in reverse order (innermost to outermost)
	// Error handling middleware is applied closest to the base transport
	if config.enableErrorHandling {
		transport = middleware.NewErrorHandlingMiddleware(transport, middleware.ErrorHandlingMiddlewareDeps{
			RootLogger: f.logger,
		})
	}

	if config.authTokenSource != nil {
		transport = &oauth2.Transport{
			Source: config.authTokenSource,
			Base:   transport,
		}
	}

	// Logging middleware is outermost to capture full request lifecycle
	if config.enableLogging {
		transport = middleware.NewLoggingMiddleware(transport, middleware.LoggingMiddlewareDeps{
			RootLogger: f.logger,
		})
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.timeout,
	}
}
