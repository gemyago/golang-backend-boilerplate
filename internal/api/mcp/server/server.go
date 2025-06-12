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
	Name         string `name:"config.mcpServer.name"`
	Version      string `name:"config.mcpServer.version"`
	StdioEnabled bool   `name:"config.mcpServer.stdioEnabled"`
	HTTPEnabled  bool   `name:"config.mcpServer.httpEnabled"`
	HTTPHost     string `name:"config.mcpServer.httpHost"`
	HTTPPort     int    `name:"config.mcpServer.httpPort"`

	// services
	*services.ShutdownHooks

	// controllers
	ControllersRegistry *controllers.Registry
}

// ToolHandler represents a function that handles tool calls.
type ToolHandler = server.ToolHandlerFunc

// ResourceHandler represents a function that handles resource reads.
type ResourceHandler = server.ResourceHandlerFunc

// ToolInfo contains information about a registered tool.
type ToolInfo struct {
	Tool    mcp.Tool
	Handler ToolHandler
}

// ResourceInfo contains information about a registered resource.
type ResourceInfo struct {
	Resource mcp.Resource
	Handler  ResourceHandler
}

// MCPServer wraps the mcp-go server with additional functionality.
type MCPServer struct {
	mcpServer   *server.MCPServer
	deps        MCPServerDeps
	logger      *slog.Logger
	initialized bool

	// HTTP server lifecycle management
	httpServer *http.Server
	sseServer  *server.SSEServer

	// Tool and resource registry
	tools     map[string]ToolInfo
	resources map[string]ResourceInfo
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
		resources:   make(map[string]ResourceInfo),
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

	// Register all tools and resources with the underlying mcp-go server
	s.registerToolsWithMCPServer()
	s.registerResourcesWithMCPServer()

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

	// Create SSE server
	s.sseServer = server.NewSSEServer(s.mcpServer)

	// Set up HTTP server
	address := fmt.Sprintf("%s:%d", s.deps.HTTPHost, s.deps.HTTPPort)

	s.httpServer = &http.Server{
		Addr:         address,
		Handler:      s.sseServer,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}

	s.logger.InfoContext(ctx, "MCP HTTP server configured",
		slog.String("address", address))

	// Start server in a goroutine so we can handle shutdown
	serverErr := make(chan error, 1)
	go func() {
		s.logger.InfoContext(ctx, "Starting HTTP server")
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
	s.sseServer = nil
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

// RegisterResource registers a resource with the MCP server.
func (s *MCPServer) RegisterResource(resource mcp.Resource, handler ResourceHandler) error {
	if s.initialized {
		return fmt.Errorf("cannot register resource after server initialization: %w", ErrServerAlreadyInitialized)
	}

	s.logger.Debug("Registering resource",
		slog.String("name", resource.Name),
		slog.String("uri", resource.URI))

	s.resources[resource.Name] = ResourceInfo{
		Resource: resource,
		Handler:  handler,
	}

	return nil
}

// GetRegisteredTools returns a list of all registered tools.
func (s *MCPServer) GetRegisteredTools() []mcp.Tool {
	tools := make([]mcp.Tool, 0, len(s.tools))
	for _, toolInfo := range s.tools {
		tools = append(tools, toolInfo.Tool)
	}
	return tools
}

// GetRegisteredResources returns a list of all registered resources.
func (s *MCPServer) GetRegisteredResources() []mcp.Resource {
	resources := make([]mcp.Resource, 0, len(s.resources))
	for _, resourceInfo := range s.resources {
		resources = append(resources, resourceInfo.Resource)
	}
	return resources
}

// GetToolByName returns a tool by its name.
func (s *MCPServer) GetToolByName(name string) (ToolInfo, bool) {
	toolInfo, exists := s.tools[name]
	return toolInfo, exists
}

// GetResourceByName returns a resource by its name.
func (s *MCPServer) GetResourceByName(name string) (ResourceInfo, bool) {
	resourceInfo, exists := s.resources[name]
	return resourceInfo, exists
}

// DiscoverTools returns information about available tools for discovery.
func (s *MCPServer) DiscoverTools() map[string]ToolInfo {
	discovered := make(map[string]ToolInfo)
	for name, toolInfo := range s.tools {
		discovered[name] = toolInfo
	}
	return discovered
}

// DiscoverResources returns information about available resources for discovery.
func (s *MCPServer) DiscoverResources() map[string]ResourceInfo {
	discovered := make(map[string]ResourceInfo)
	for name, resourceInfo := range s.resources {
		discovered[name] = resourceInfo
	}
	return discovered
}

// registerToolsWithMCPServer registers all tools with the underlying mcp-go server.
func (s *MCPServer) registerToolsWithMCPServer() {
	for _, toolInfo := range s.tools {
		s.mcpServer.AddTool(toolInfo.Tool, toolInfo.Handler)
		s.logger.Debug("Registered tool with MCP server", slog.String("name", toolInfo.Tool.Name))
	}
}

// registerResourcesWithMCPServer registers all resources with the underlying mcp-go server.
func (s *MCPServer) registerResourcesWithMCPServer() {
	for _, resourceInfo := range s.resources {
		s.mcpServer.AddResource(resourceInfo.Resource, resourceInfo.Handler)
		s.logger.Debug("Registered resource with MCP server", slog.String("name", resourceInfo.Resource.Name))
	}
}
