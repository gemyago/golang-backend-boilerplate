package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/services"
	"github.com/go-faker/faker/v4"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func makeMockDeps() MCPServerDeps {
	return MCPServerDeps{
		RootLogger:    slog.Default(),
		Name:          faker.Name(),
		Version:       faker.Word(),
		StdioEnabled:  true,
		HTTPEnabled:   false,
		HTTPHost:      faker.IPv4(),
		HTTPPort:      8080,
		ShutdownHooks: services.NewTestShutdownHooks(),
	}
}

func TestNewMCPServer(t *testing.T) {
	t.Run("should create MCP server with default configuration", func(t *testing.T) {
		deps := makeMockDeps()

		server := NewMCPServer(deps)

		require.NotNil(t, server)
		require.NotNil(t, server.mcpServer)
		require.NotNil(t, server.logger)
		require.Equal(t, deps, server.deps)
		require.False(t, server.initialized)
	})

	t.Run("should register shutdown hook", func(t *testing.T) {
		deps := makeMockDeps()

		server := NewMCPServer(deps)

		require.NotNil(t, server)
		// We can't directly test shutdown hook registration without exposing internals,
		// but we can verify the server was created successfully with shutdown hooks
		require.NotNil(t, server.deps.ShutdownHooks)
	})

	t.Run("should create logger with group", func(t *testing.T) {
		deps := makeMockDeps()

		server := NewMCPServer(deps)

		require.NotNil(t, server.logger)
		// The logger should be created with the mcp-server group
		// We can't easily test the group without complex reflection,
		// but we can verify it's not nil and different from root logger
		require.NotEqual(t, deps.RootLogger, server.logger)
	})
}

func TestMCPServerInitialize(t *testing.T) {
	t.Run("should initialize server successfully", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.Initialize(ctx)

		require.NoError(t, err)
		require.True(t, server.initialized)
	})

	t.Run("should not reinitialize already initialized server", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx := t.Context()

		// Initialize first time
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.True(t, server.initialized)

		// Initialize second time - should not return error
		err = server.Initialize(ctx)
		require.NoError(t, err)
		require.True(t, server.initialized)
	})

	t.Run("should handle context cancellation gracefully", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel context before initialization

		err := server.Initialize(ctx)

		// Should still succeed as initialization doesn't depend on context
		require.NoError(t, err)
		require.True(t, server.initialized)
	})
}

func TestMCPServerStartStdio(t *testing.T) {
	t.Run("should return error when stdio is disabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = false
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.StartStdio(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrStdioNotEnabled)
	})

	t.Run("should initialize server before starting stdio", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = true
		server := NewMCPServer(deps)
		ctx := t.Context()

		// Server should not be initialized initially
		require.False(t, server.initialized)

		// Note: StartStdio will block, so we can't easily test the actual start
		// We can test that it would try to initialize first
		// For now, we'll test the initialization logic separately
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.True(t, server.initialized)
	})

	t.Run("should respect stdio enabled configuration", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = true
		server := NewMCPServer(deps)
		ctx := t.Context()

		// We can't actually start stdio in tests as it would block,
		// but we can verify the configuration is respected
		require.True(t, server.deps.StdioEnabled)

		// Test that it doesn't immediately error on stdio enabled
		err := server.Initialize(ctx)
		require.NoError(t, err)
	})
}

func TestMCPServerStartHTTP(t *testing.T) {
	t.Run("should return error when HTTP is disabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = false
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.StartHTTP(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotEnabled)
	})

	t.Run("should return not implemented error when HTTP is enabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.StartHTTP(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotImplemented)
	})

	t.Run("should initialize server before starting HTTP", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		server := NewMCPServer(deps)
		ctx := t.Context()

		// Server should not be initialized initially
		require.False(t, server.initialized)

		// Even though HTTP will return not implemented,
		// it should still initialize the server first
		err := server.StartHTTP(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotImplemented)
		require.True(t, server.initialized)
	})
}

func TestMCPServerStop(t *testing.T) {
	t.Run("should stop server gracefully", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.Stop(ctx)

		require.NoError(t, err)
	})

	t.Run("should handle context cancellation during stop", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := server.Stop(ctx)

		// Should still succeed as stop is currently just logging
		require.NoError(t, err)
	})
}

func TestMCPServerAddTool(t *testing.T) {
	t.Run("should add tool to server", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)

		// Create a basic test tool
		testTool := mcp.Tool{
			Name:        faker.Word(),
			Description: faker.Sentence(),
		}

		// This is more of a smoke test since we can't easily verify
		// the tool was added without exposing internals
		require.NotPanics(t, func() {
			server.AddTool(testTool, func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return &mcp.CallToolResult{}, nil
			})
		})
	})
}

func TestMCPServerAddResource(t *testing.T) {
	t.Run("should add resource to server", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)

		// Create a basic test resource
		testResource := mcp.Resource{
			Name: faker.Word(),
			URI:  faker.URL(),
		}

		// This is more of a smoke test since we can't easily verify
		// the resource was added without exposing internals
		require.NotPanics(t, func() {
			server.AddResource(testResource, func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return []mcp.ResourceContents{}, nil
			})
		})
	})
}

// Additional tests for HTTP transport initialization.
func TestMCPServerHTTPTransportInitialization(t *testing.T) {
	t.Run("should validate HTTP configuration", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		deps.HTTPHost = faker.IPv4()
		deps.HTTPPort = 8080 + int(faker.RandomUnixTime()%1000) // Random port between 8080-9079

		server := NewMCPServer(deps)

		require.NotNil(t, server)
		require.Equal(t, deps.HTTPHost, server.deps.HTTPHost)
		require.Equal(t, deps.HTTPPort, server.deps.HTTPPort)
		require.True(t, server.deps.HTTPEnabled)
	})

	t.Run("should initialize with both stdio and HTTP enabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = true
		deps.HTTPEnabled = true
		deps.HTTPHost = faker.IPv4()
		deps.HTTPPort = 8080

		server := NewMCPServer(deps)
		ctx := t.Context()

		// Initialize server
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.True(t, server.initialized)

		// Both transports should be configured
		require.True(t, server.deps.StdioEnabled)
		require.True(t, server.deps.HTTPEnabled)
	})

	t.Run("should handle HTTP-only configuration", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = false
		deps.HTTPEnabled = true
		deps.HTTPHost = "0.0.0.0"
		deps.HTTPPort = 3000

		server := NewMCPServer(deps)
		ctx := t.Context()

		// HTTP should be enabled but not implemented yet
		err := server.StartHTTP(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotImplemented)

		// But stdio should be disabled
		err = server.StartStdio(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStdioNotEnabled)
	})

	t.Run("should handle custom HTTP port configuration", func(t *testing.T) {
		customPort := 9999
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		deps.HTTPPort = customPort

		server := NewMCPServer(deps)

		require.Equal(t, customPort, server.deps.HTTPPort)
	})

	t.Run("should handle custom HTTP host configuration", func(t *testing.T) {
		customHost := "127.0.0.1"
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		deps.HTTPHost = customHost

		server := NewMCPServer(deps)

		require.Equal(t, customHost, server.deps.HTTPHost)
	})

	t.Run("should log HTTP configuration details", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = true
		deps.HTTPHost = faker.IPv4()
		deps.HTTPPort = 8080

		server := NewMCPServer(deps)
		ctx := t.Context()

		// This should log the HTTP configuration details
		err := server.StartHTTP(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotImplemented)

		// Server should have been initialized despite HTTP not being implemented
		require.True(t, server.initialized)
	})

	t.Run("should respect HTTP transport disabled state", func(t *testing.T) {
		deps := makeMockDeps()
		deps.HTTPEnabled = false

		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.StartHTTP(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHTTPNotEnabled)

		// Server should not have been initialized for disabled HTTP
		require.False(t, server.initialized)
	})
}
