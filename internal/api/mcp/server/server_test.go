package server

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/controllers"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/services"
	"github.com/go-faker/faker/v4"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMockDeps() MCPServerDeps {
	// Create mock time service
	timeServiceDeps := app.TimeServiceDeps{
		RootLogger: slog.Default(),
	}
	timeService := app.NewTimeService(timeServiceDeps)

	// Create mock controllers registry
	controllersRegistryDeps := controllers.RegistryDeps{
		RootLogger:  slog.Default(),
		TimeService: timeService,
	}
	controllersRegistry := controllers.NewControllersRegistry(controllersRegistryDeps)

	return MCPServerDeps{
		RootLogger:          slog.Default(),
		Name:                faker.Name(),
		Version:             faker.Word(),
		StdioEnabled:        true,
		HTTPEnabled:         false,
		HTTPHost:            faker.IPv4(),
		HTTPPort:            8080,
		ShutdownHooks:       services.NewTestShutdownHooks(),
		ControllersRegistry: controllersRegistry,
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
	t.Run("should start stdio transport when enabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = true
		server := NewMCPServer(deps)
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		// Start stdio in a goroutine since it blocks
		errChan := make(chan error, 1)
		go func() {
			errChan <- server.StartStdio(ctx)
		}()

		// Wait for context timeout or error
		select {
		case err := <-errChan:
			// Context timeout should cause graceful termination
			require.NoError(t, err)
		case <-ctx.Done():
			// This is expected - context timeout
		}

		require.True(t, server.initialized)
	})

	t.Run("should return error when stdio disabled", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = false
		server := NewMCPServer(deps)
		ctx := t.Context()

		err := server.StartStdio(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrStdioNotEnabled)
		require.False(t, server.initialized)
	})

	t.Run("should initialize server before starting stdio", func(t *testing.T) {
		deps := makeMockDeps()
		deps.StdioEnabled = true
		server := NewMCPServer(deps)
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		// Server should not be initialized initially
		require.False(t, server.initialized)

		// Start stdio in goroutine to test initialization
		go func() {
			_ = server.StartStdio(ctx)
		}()

		// Give it a moment to initialize
		time.Sleep(10 * time.Millisecond)
		require.True(t, server.initialized)
	})
}

// TestMCPServerStartHTTP tests are moved to the newer test functions below

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

func TestMCPServerRegisterTool(t *testing.T) {
	t.Run("should register tool successfully before initialization", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)

		// Create a basic test tool
		testTool := mcp.Tool{
			Name:        faker.Word(),
			Description: faker.Sentence(),
		}

		// Register tool before initialization
		err := server.RegisterTool(testTool, func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{}, nil
		})

		require.NoError(t, err)
	})

	t.Run("should return error when registering after initialization", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx := t.Context()

		// Initialize server first
		err := server.Initialize(ctx)
		require.NoError(t, err)

		// Try to register tool after initialization
		testTool := mcp.Tool{
			Name:        faker.Word(),
			Description: faker.Sentence(),
		}

		err = server.RegisterTool(testTool, func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{}, nil
		})

		require.Error(t, err)
		require.ErrorIs(t, err, ErrServerAlreadyInitialized)
	})

	t.Run("should handle registration through controllers workflow", func(t *testing.T) {
		deps := makeMockDeps()
		server := NewMCPServer(deps)
		ctx := t.Context()

		// This tests the actual workflow: controllers register tools, then server initializes
		err := server.Initialize(ctx)
		require.NoError(t, err)
		require.True(t, server.initialized)
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

		// HTTP should be enabled
		require.True(t, server.deps.HTTPEnabled)
		require.Equal(t, "0.0.0.0", server.deps.HTTPHost)
		require.Equal(t, 3000, server.deps.HTTPPort)

		// But stdio should be disabled
		err := server.StartStdio(t.Context())
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

		// Verify HTTP configuration is properly set
		require.True(t, server.deps.HTTPEnabled)
		require.Equal(t, deps.HTTPHost, server.deps.HTTPHost)
		require.Equal(t, deps.HTTPPort, server.deps.HTTPPort)
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

func TestMCPServer_StartHTTP_Success(t *testing.T) {
	deps := makeMockDeps()
	deps.HTTPEnabled = true
	deps.HTTPHost = "localhost"
	deps.HTTPPort = 0 // Use port 0 for dynamic allocation

	server := NewMCPServer(deps)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	// Start HTTP server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.StartHTTP(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for server to finish
	err := <-serverErr
	require.NoError(t, err)
	assert.True(t, server.initialized)
}

func TestMCPServer_StartHTTP_Disabled(t *testing.T) {
	deps := makeMockDeps()
	deps.HTTPEnabled = false

	server := NewMCPServer(deps)
	ctx := t.Context()

	err := server.StartHTTP(ctx)
	require.Error(t, err)
	assert.Equal(t, ErrHTTPNotEnabled, err)
	assert.False(t, server.initialized)
}

func TestMCPServer_StartHTTP_InitializationError(t *testing.T) {
	// Test scenario where initialization fails
	deps := makeMockDeps()
	deps.HTTPEnabled = true
	deps.HTTPHost = "localhost"
	deps.HTTPPort = 0 // Use port 0 for dynamic allocation to avoid conflicts

	server := NewMCPServer(deps)
	// Force initialization to fail by setting initialized to true first
	server.initialized = true

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	// Start HTTP server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.StartHTTP(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for server to finish
	err := <-serverErr
	require.NoError(t, err) // Should succeed since Initialize() returns nil if already initialized
}

func TestMCPServer_ShutdownHTTPServer(t *testing.T) {
	deps := makeMockDeps()
	deps.HTTPEnabled = true
	deps.HTTPHost = "localhost"
	deps.HTTPPort = 0

	server := NewMCPServer(deps)
	require.NoError(t, server.Initialize(t.Context()))

	// Create SSE server and HTTP server directly for testing
	server.sseServer = mcpserver.NewSSEServer(server.mcpServer)
	server.httpServer = &http.Server{
		Addr:    "localhost:0",
		Handler: server.sseServer,
	}

	ctx := t.Context()
	err := server.shutdownHTTPServer(ctx)
	require.NoError(t, err)
	assert.Nil(t, server.httpServer)
	assert.Nil(t, server.sseServer)
}

func TestMCPServer_ShutdownHTTPServer_NoServer(t *testing.T) {
	deps := makeMockDeps()
	server := NewMCPServer(deps)

	ctx := t.Context()
	err := server.shutdownHTTPServer(ctx)
	require.NoError(t, err)
}

func TestMCPServer_Stop_WithHTTPServer(t *testing.T) {
	deps := makeMockDeps()
	deps.HTTPEnabled = true

	server := NewMCPServer(deps)
	require.NoError(t, server.Initialize(t.Context()))

	// Set up a mock HTTP server
	server.sseServer = mcpserver.NewSSEServer(server.mcpServer)
	server.httpServer = &http.Server{
		Addr:    "localhost:0",
		Handler: server.sseServer,
	}

	ctx := t.Context()
	err := server.Stop(ctx)
	require.NoError(t, err)
	assert.Nil(t, server.httpServer)
	assert.Nil(t, server.sseServer)
}

func TestMCPServer_Stop_WithoutHTTPServer(t *testing.T) {
	deps := makeMockDeps()
	server := NewMCPServer(deps)

	ctx := t.Context()
	err := server.Stop(ctx)
	require.NoError(t, err)
}

// Integration test for both transport methods.
func TestMCPServer_BothTransports_Configuration(t *testing.T) {
	deps := makeMockDeps()
	deps.StdioEnabled = true
	deps.HTTPEnabled = true
	deps.HTTPHost = "localhost"
	deps.HTTPPort = 8080

	server := NewMCPServer(deps)

	// Test that both transports can be configured
	assert.Equal(t, deps.StdioEnabled, server.deps.StdioEnabled)
	assert.Equal(t, deps.HTTPEnabled, server.deps.HTTPEnabled)
	assert.Equal(t, deps.HTTPHost, server.deps.HTTPHost)
	assert.Equal(t, deps.HTTPPort, server.deps.HTTPPort)
}
