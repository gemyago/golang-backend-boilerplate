package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/services"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"
)

// MCPServerDeps contains dependencies for creating the MCP server.
type MCPServerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Name         string `name:"config.mcpServer.name"`
	Version      string `name:"config.mcpServer.version"`
	StdioEnabled bool   `name:"config.mcpServer.stdioEnabled"`
	HTTPEnabled  bool   `name:"config.mcpServer.httpEnabled"`
	HTTPHost     string `name:"config.mcpServer.httpHost"`
	HTTPPort     int    `name:"config.mcpServer.httpPort"`

	// services
	*services.ShutdownHooks
}

// MCPServer wraps the mcp-go server with additional functionality.
type MCPServer struct {
	mcpServer   *server.MCPServer
	deps        MCPServerDeps
	logger      *slog.Logger
	initialized bool
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

	// TODO: Add tools and resources registration in later tasks
	// This is where we'll register time and math tools

	s.initialized = true
	s.logger.InfoContext(ctx, "MCP server initialized successfully")

	return nil
}

// StartStdio starts the MCP server with stdio transport.
func (s *MCPServer) StartStdio(ctx context.Context) error {
	if !s.deps.StdioEnabled {
		s.logger.WarnContext(ctx, "Attempted to start stdio transport but it is disabled")
		return ErrStdioNotEnabled
	}

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
	if !s.deps.HTTPEnabled {
		return ErrHTTPNotEnabled
	}

	if err := s.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP server for HTTP transport: %w", err)
	}

	s.logger.InfoContext(ctx, "Starting MCP server with HTTP transport",
		slog.String("host", s.deps.HTTPHost),
		slog.Int("port", s.deps.HTTPPort))

	// TODO: Implement HTTP server in task 2.5
	// For now, just return an error indicating it's not implemented
	return ErrHTTPNotImplemented
}

// Stop gracefully stops the MCP server.
func (s *MCPServer) Stop(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Stopping MCP server")

	// TODO: Implement graceful shutdown in task 2.6
	// For now, just log that we're stopping

	return nil
}

// AddTool adds a tool to the MCP server.
func (s *MCPServer) AddTool(
	tool mcp.Tool,
	handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error),
) {
	s.mcpServer.AddTool(tool, handler)
	s.logger.Debug("Added tool to MCP server", "tool", tool.Name)
}

// AddResource adds a resource to the MCP server.
func (s *MCPServer) AddResource(
	resource mcp.Resource,
	handler func(context.Context, mcp.ReadResourceRequest) ([]mcp.ResourceContents, error),
) {
	s.mcpServer.AddResource(resource, handler)
	s.logger.Debug("Added resource to MCP server", "resource", resource.Name)
}
