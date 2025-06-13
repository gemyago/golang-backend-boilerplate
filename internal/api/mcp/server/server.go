package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/controllers"
	"github.com/gemyago/golang-backend-boilerplate/internal/services"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"
)

// Constants for server configuration.
const (
	httpReadTimeout  = 30 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
	shutdownTimeout  = 10 * time.Second
)

// MCPServerDeps contains dependencies for creating the MCP server.
type MCPServerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Name     string `name:"config.mcpServer.name"`
	Version  string `name:"config.mcpServer.version"`
	HTTPHost string `name:"config.mcpServer.httpHost"`
	HTTPPort int    `name:"config.mcpServer.httpPort"`

	// services
	*services.ShutdownHooks

	// controllers
	ControllersRegistry *controllers.Registry
}

// ToolHandler represents a function that handles tool calls.
type ToolHandler = server.ToolHandlerFunc

// ToolInfo contains information about a registered tool.
type ToolInfo struct {
	Tool    mcp.Tool
	Handler ToolHandler
}

// MCPServer wraps the mcp-go server with additional functionality.
type MCPServer struct {
	mcpServer   *server.MCPServer
	deps        MCPServerDeps
	logger      *slog.Logger
	initialized bool

	httpServer *http.Server

	// Tool registry
	tools map[string]ToolInfo
}

// NewMCPServer creates a new MCP server instance.
func NewMCPServer(deps MCPServerDeps) *MCPServer {
	// Create the underlying mcp-go server
	mcpServer := server.NewMCPServer(
		deps.Name,
		deps.Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	mcpSrv := &MCPServer{
		deps:        deps,
		mcpServer:   mcpServer,
		logger:      deps.RootLogger.WithGroup("mcp-server"),
		initialized: false,
		tools:       make(map[string]ToolInfo),
	}

	// Register shutdown hook
	deps.ShutdownHooks.Register("mcp-server", mcpSrv.Stop)

	return mcpSrv
}

// Initialize sets up the MCP server with tools and resources.
func (s *MCPServer) Initialize(ctx context.Context) error {
	if s.initialized {
		return nil
	}

	s.logger.InfoContext(ctx, "Initializing MCP server",
		slog.String("name", s.deps.Name),
		slog.String("version", s.deps.Version))

	// Register all controllers with the MCP server
	if err := s.deps.ControllersRegistry.RegisterAllControllers(ctx, s); err != nil {
		return fmt.Errorf("failed to register MCP controllers: %w", err)
	}

	// Register all tools with the underlying mcp-go server
	s.registerToolsWithMCPServer()

	s.initialized = true
	s.logger.InfoContext(ctx, "MCP server initialized successfully")

	return nil
}

// StartStdio starts the MCP server with stdio transport.
func (s *MCPServer) StartStdio(ctx context.Context) error {
	if err := s.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP server for stdio transport: %w", err)
	}

	s.logger.InfoContext(ctx, "Starting MCP server with stdio transport",
		slog.String("name", s.deps.Name),
		slog.String("version", s.deps.Version))

	// Start the stdio server - this will block until the connection is closed
	// The mcp-go framework handles all the protocol details
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("MCP stdio server terminated with error: %w", err)
	}

	s.logger.InfoContext(ctx, "MCP stdio server terminated gracefully")
	return nil
}

// StartHTTP starts the MCP server with HTTP transport.
func (s *MCPServer) StartHTTP(ctx context.Context) error {
	if err := s.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP server for HTTP transport: %w", err)
	}

	// Set up HTTP server
	address := fmt.Sprintf("%s:%d", s.deps.HTTPHost, s.deps.HTTPPort)

	s.httpServer = &http.Server{
		Addr:         address,
		Handler:      server.NewSSEServer(s.mcpServer),
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}

	// Start server in a goroutine so we can handle shutdown
	serverErr := make(chan error, 1)
	go func() {
		s.logger.InfoContext(ctx, "Starting MCP server with HTTP transport",
			slog.String("host", s.deps.HTTPHost),
			slog.Int("port", s.deps.HTTPPort))
		serverErr <- s.httpServer.ListenAndServe()
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.InfoContext(ctx, "Context cancelled, shutting down HTTP server")
		return s.shutdownHTTPServer(context.Background())
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		s.logger.InfoContext(ctx, "HTTP server terminated gracefully")
		return nil
	}
}

// shutdownHTTPServer gracefully shuts down the HTTP server.
func (s *MCPServer) shutdownHTTPServer(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.InfoContext(ctx, "Shutting down HTTP server")

	// Give the server 10 seconds to finish serving connections
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to gracefully shutdown HTTP server",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.logger.InfoContext(ctx, "HTTP server shutdown completed")
	s.httpServer = nil
	return nil
}

// Stop gracefully stops the MCP server.
func (s *MCPServer) Stop(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Stopping MCP server")

	// Shutdown HTTP server if it's running
	if s.httpServer != nil {
		if err := s.shutdownHTTPServer(ctx); err != nil {
			s.logger.ErrorContext(ctx, "Error shutting down HTTP server",
				slog.String("error", err.Error()))
			// Don't return the error, continue with cleanup
		}
	}

	s.logger.InfoContext(ctx, "MCP server stopped successfully")
	return nil
}

// RegisterTool registers a tool with the MCP server.
func (s *MCPServer) RegisterTool(tool mcp.Tool, handler ToolHandler) error {
	if s.initialized {
		return fmt.Errorf("cannot register tool after server initialization: %w", ErrServerAlreadyInitialized)
	}

	s.logger.Debug("Registering tool",
		slog.String("name", tool.Name),
		slog.String("description", tool.Description))

	s.tools[tool.Name] = ToolInfo{
		Tool:    tool,
		Handler: handler,
	}

	return nil
}

// registerToolsWithMCPServer registers all tools with the underlying mcp-go server.
func (s *MCPServer) registerToolsWithMCPServer() {
	for _, toolInfo := range s.tools {
		s.mcpServer.AddTool(toolInfo.Tool, toolInfo.Handler)
		s.logger.Debug("Registered tool with MCP server", slog.String("name", toolInfo.Tool.Name))
	}
}
