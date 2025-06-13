package server

import "errors"

// Error definitions for MCP server.
var (
	// ErrServerAlreadyInitialized is returned when trying to initialize an already initialized server.
	ErrServerAlreadyInitialized = errors.New("MCP server is already initialized")
)
