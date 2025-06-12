package server

import "errors"

// Error definitions for MCP server.
var (
	// ErrStdioNotEnabled is returned when trying to start stdio transport when it's disabled.
	ErrStdioNotEnabled = errors.New("stdio transport is not enabled")

	// ErrHTTPNotEnabled is returned when trying to start HTTP transport when it's disabled.
	ErrHTTPNotEnabled = errors.New("HTTP transport is not enabled")

	// ErrServerNotInitialized is returned when trying to use server before initialization.
	ErrServerNotInitialized = errors.New("MCP server is not initialized")

	// ErrServerAlreadyInitialized is returned when trying to initialize an already initialized server.
	ErrServerAlreadyInitialized = errors.New("MCP server is already initialized")
)
