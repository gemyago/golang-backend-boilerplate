# PRD: MCP Tools Extension for Golang Backend Boilerplate

## Introduction/Overview

This feature extends the existing golang backend boilerplate to support MCP (Model Context Protocol) tools for AI agents. The extension will allow developers to quickly create and deploy MCP servers that can integrate with AI agents like Claude Desktop through standardized protocols. The goal is to provide a simple, consistent way to build MCP tools that follow the existing boilerplate's architectural patterns and conventions.

## Goals

1. **Rapid MCP Tool Development**: Enable developers to create MCP tools quickly using familiar patterns from the existing boilerplate
2. **Multiple Transport Support**: Provide both stdio and HTTP transport methods for maximum compatibility
3. **Framework Integration**: Leverage existing Go MCP frameworks for auto-discovery and reduced boilerplate code
4. **Configuration Consistency**: Use the existing project structure to handle configuration for MCP tools (./config/*json files and ./config/provide.go to inject specific config values)
5. **Developer Experience**: Maintain the same development workflow and patterns as the existing server and jobs commands

## User Stories

1. **As a developer**, I want to start an MCP server using `cmd/mcp stdio` so that I can integrate with Claude Desktop through stdio transport
2. **As a developer**, I want to start an MCP server using `cmd/mcp http` so that I can integrate with web-based AI agents through HTTP transport  
3. **As a developer**, I want to create new MCP tools following existing patterns so that I can maintain consistency with the boilerplate architecture
4. **As a developer**, I want MCP tools to be auto-discovered so that I don't need to manually register each tool
5. **As a developer**, I want to configure MCP settings using the existing config system so that I can manage all application settings in one place
6. **As an AI agent user**, I want to call simple mathematical operations through MCP so that I can perform calculations
7. **As an AI agent user**, I want to get current time information through MCP so that I can access time-based data

## Functional Requirements

### FR1: Command Structure
The system must provide a new `cmd/mcp` command with the following subcommands:
- `mcp stdio` - Start MCP server with stdio transport
- `mcp http` - Start MCP server with HTTP transport

### FR2: Tool Implementation
The system must implement the following initial MCP tools:
- **time-tool**: Returns current date/time in various formats
- **math-tool**: Performs basic arithmetic operations (add, subtract, multiply, divide)

### FR3: Framework Integration
The system must integrate with the `mark3labs/mcp-go` framework to provide:
- Automatic tool discovery and registration
- Type-safe parameter handling
- JSON schema generation for tool parameters
- Standard MCP protocol compliance

### FR4: Configuration Management
The system must extend the existing configuration system to support:
- MCP server settings (port, host for HTTP transport)
- Tool-specific configuration options (may not be needed for simple tools)
- Transport-specific settings

### FR5: Architecture Consistency
The system must follow existing boilerplate patterns:
- Use cobra CLI framework for command structure
- Implement dependency injection using dig container
- Follow the same logging patterns with slog
- Maintain the same project structure conventions

### FR6: Transport Support
The system must support both transport methods:
- **stdio**: For integration with Claude Desktop and similar clients
- **HTTP**: For web-based integrations and development/testing

## Non-Goals (Out of Scope)

1. **Complex Tools**: Advanced tools requiring database integration or external API calls
2. **Resource Implementation**: MCP resources functionality (will be future enhancement)
3. **Prompt Implementation**: MCP prompts functionality (will be future enhancement)
4. **Authentication**: User authentication or authorization mechanisms
5. **Rate Limiting**: Request throttling or rate limiting features
6. **Tool Marketplace**: Plugin system or external tool discovery
7. **GUI Interface**: Web-based administration or management interface

## Design Considerations

### Framework Choice
- **Primary**: Use `mark3labs/mcp-go` framework for its comprehensive feature set and auto-discovery capabilities
- **Fallback**: If auto-discovery is not available, implement static tool registration
- **Type Safety**: Leverage Go's type system with automatic JSON schema generation

### Project Structure
```
cmd/
  mcp/
    main.go
    root.go
    stdio.go
    http.go
internal/
  api/mcp/
    -- adapter level between protocol and application logic
    controllers/
      time.go
      math.go
    server/
      -- register controllers
      server.go

  -- application level
  app/
    time.go
    math.go
  services/ -- infrastructure level, may not be needed for simple tools
    time.go
    math.go
```

## Technical Considerations

### Dependencies
- Add `github.com/mark3labs/mcp-go` as primary MCP framework dependency
- Maintain existing dependencies (cobra, viper, dig, slog)
- Consider version compatibility with existing Go modules

### Tool Registration
- Use reflection-based auto-discovery if supported by the framework
- Implement tool interface for consistent tool development

### Error Handling
- Follow existing error handling patterns
- Provide meaningful error messages for MCP protocol violations
- Implement graceful degradation for tool failures

### Testing Strategy
- Unit tests for individual tools
- Integration tests for MCP protocol compliance
- End-to-end tests with actual MCP clients - can be done with scripts or manually

## Success Metrics

1. **Development Speed**: Developers can create and deploy a basic MCP server in under 10 minutes
2. **Tool Creation**: New MCP tools can be added with less than 50 lines of code
3. **Integration Success**: Successfully integrates with Claude Desktop or Cursor on first attempt
4. **Performance**: MCP server responds to tool calls within 100ms for simple operations
5. **Reliability**: Zero protocol compliance issues when tested with standard MCP clients

## Open Questions

1. **Configuration Location**: Should MCP configuration be in the main config file or separate mcp.yaml? - config/default.json and config/provide.go
2. **Tool Interface**: What should the standard interface look like for custom tools?
3. **Error Formats**: How should tool errors be formatted and returned to clients?
4. **Logging**: Should MCP operations use separate log levels or follow existing patterns?
5. **Hot Reload**: Should the server support hot reloading of tool configuration changes? - no
6. **Versioning**: How should tool versions be managed and communicated to clients? - out of scope
7. **Testing Framework**: Should we create MCP-specific test utilities or use standard testing approaches? - out of scope