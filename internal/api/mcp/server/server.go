package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	httpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
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

type ToolController interface {
	NewTools() []server.ServerTool
}

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
	Controllers []ToolController
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
	mcpServer     *server.MCPServer
	deps          MCPServerDeps
	logger        *slog.Logger
	shutdownHooks *services.ShutdownHooks
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
		deps:          deps,
		mcpServer:     mcpServer,
		logger:        deps.RootLogger.WithGroup("mcp-server"),
		shutdownHooks: deps.ShutdownHooks,
	}

	for _, controller := range deps.Controllers {
		tools := controller.NewTools()
		mcpSrv.mcpServer.AddTools(tools...)
	}

	return mcpSrv
}

// StartStdio starts the MCP server with stdio transport.
func (s *MCPServer) StartStdio(ctx context.Context) error {
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
	s.logger.InfoContext(ctx, "Starting MCP server with HTTP transport",
		slog.String("name", s.deps.Name),
		slog.String("version", s.deps.Version))

	httpSrv := httpserver.NewHTTPServer(httpserver.HTTPServerDeps{
		RootLogger: s.logger,

		Host:              s.deps.HTTPHost,
		Port:              s.deps.HTTPPort,
		IdleTimeout:       httpIdleTimeout,
		ReadHeaderTimeout: httpReadTimeout,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,

		ShutdownHooks: s.shutdownHooks,
		Handler:       server.NewSSEServer(s.mcpServer),
	})

	return httpSrv.Start(ctx)
}
