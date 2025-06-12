## Relevant Files

### Source Files
- `cmd/mcp/main.go` - Entry point for MCP command
- `cmd/mcp/root.go` - Root command setup for MCP CLI
- `cmd/mcp/stdio.go` - stdio transport command implementation
- `cmd/mcp/http.go` - HTTP transport command implementation
- `internal/api/mcp/server/server.go` - MCP server initialization and setup
- `internal/api/mcp/controllers/time.go` - Time tool controller for MCP
- `internal/api/mcp/controllers/math.go` - Math tool controller for MCP
- `internal/app/time.go` - Time tool business logic
- `internal/app/math.go` - Math tool business logic
- `internal/services/time.go` - Time service infrastructure (if needed)
- `internal/services/math.go` - Math service infrastructure (if needed)
- `config/default.json` - Configuration file with MCP settings
- `config/provide.go` - Dependency injection configuration
- `internal/di/container.go` - Dependency injection container registration
- `go.mod` - Go module dependencies (add mcp-go framework)

### Test Files
- `cmd/mcp/stdio_test.go` - Unit tests for stdio command
- `cmd/mcp/http_test.go` - Unit tests for HTTP command
- `internal/api/mcp/controllers/time_test.go` - Unit tests for time controller
- `internal/api/mcp/controllers/math_test.go` - Unit tests for math controller
- `internal/app/time_test.go` - Unit tests for time business logic
- `internal/app/math_test.go` - Unit tests for math business logic
- `internal/api/mcp/server/server_test.go` - Integration tests for MCP server
- `internal/services/mocks/mock_time_service.go` - Mock implementations for testing
- `internal/services/mocks/mock_math_service.go` - Mock implementations for testing

### Configuration Files
- `go.mod` - Go module dependencies (add github.com/mark3labs/mcp-go)
- `.mockery.yaml` - Mockery configuration for generating MCP-related mocks
- `Makefile` - Build and test commands (add MCP-specific targets)

### Notes

- **MCP Framework:** Use `github.com/mark3labs/mcp-go` framework as specified in PRD for protocol implementation and auto-discovery
- **Architecture:** Follow clean architecture with layers: MCP Controllers → App → Services
- **Transport Support:** Implement both stdio and HTTP transports for maximum compatibility
- **Code Organization:** Follow existing structure with new MCP-specific packages
- **Tool Interface:** Create consistent interface for MCP tools following Go conventions
- **Configuration:** Extend existing config system with MCP-specific settings
- **Error Handling:** Use Go's idiomatic error handling with MCP protocol compliance
- **Testing Commands:**
  - `make test` - Run all tests including MCP tests
  - `go test -v ./internal/api/mcp/... --run TestName` - Run MCP-specific tests
  - `gow test -v ./internal/api/mcp/... --run TestName` - Watch mode for MCP test development
- **MCP Commands:**
  - `go run cmd/mcp/main.go stdio` - Start MCP server with stdio transport
  - `go run cmd/mcp/main.go http` - Start MCP server with HTTP transport
- **Development Workflow:** Follow TDD cycle with stub → test → implement → refactor
- **Tool Registration:** Use framework's auto-discovery for tool registration
- **Protocol Compliance:** Ensure all implementations follow MCP protocol standards
- **Mock Generation:** Generate mocks as needed during development using mockery, not as a final step

## Tasks

- [x] 1.0 Set up MCP command structure and add mark3labs/mcp-go framework dependency
  - [x] 1.1 Add `github.com/mark3labs/mcp-go` dependency to go.mod
  - [x] 1.2 Create basic `cmd/mcp/main.go` entry point with cobra CLI setup
  - [x] 1.3 Create `cmd/mcp/root.go` with root command configuration
  - [x] 1.4 Create stub `cmd/mcp/stdio.go` command handler
  - [x] 1.5 Create stub `cmd/mcp/http.go` command handler
  - [x] 1.6 Write unit tests for command parsing and basic functionality
  - [x] 1.7 Verify commands are properly registered and accessible via CLI

- [x] 2.0 Implement MCP server infrastructure with stdio/HTTP transport support using mcp-go framework
  - [x] 2.1 Create `internal/api/mcp/server/server.go` with stub MCP server struct
  - [x] 2.2 Write test for server initialization with stdio transport
  - [x] 2.3 Implement stdio transport server initialization using mcp-go
  - [x] 2.4 Write test for server initialization with HTTP transport
  - [x] 2.5 Implement HTTP transport server initialization using mcp-go
  - [x] 2.6 Add server lifecycle management (start, stop, graceful shutdown)
  - [x] 2.7 Write integration tests for both transport methods
  - [x] 2.8 Implement tool discovery and registration mechanism
  - [x] 2.9 Add error handling and logging for server operations

- [x] 3.0 Create time tool implementation following mcp-go patterns
  - [x] 3.1 Create stub `internal/app/time.go` with time service interface
  - [x] 3.2 Write test for getting current time in ISO format
  - [x] 3.3 Implement current time functionality in app layer
  - [x] 3.4 Write test for getting current time in different formats (Unix, RFC3339)
  - [x] 3.5 Implement multiple time format support
  - [x] 3.6 Create `internal/api/mcp/controllers/time.go` MCP controller stub
  - [x] 3.7 Write test for MCP time tool registration and parameter handling
  - [x] 3.8 Implement MCP time tool controller with mcp-go framework
  - [x] 3.9 Write test for time tool execution through MCP protocol
  - [x] 3.10 Test integration with MCP server and verify JSON schema generation

- [ ] 4.0 Create math tool implementation following mcp-go patterns
  - [ ] 4.1 Create stub `internal/app/math.go` with math service interface
  - [ ] 4.2 Write test for basic addition operation
  - [ ] 4.3 Implement addition functionality in app layer
  - [ ] 4.4 Write test for subtraction operation
  - [ ] 4.5 Implement subtraction functionality
  - [ ] 4.6 Write test for multiplication operation
  - [ ] 4.7 Implement multiplication functionality
  - [ ] 4.8 Write test for division operation with error handling
  - [ ] 4.9 Implement division functionality with zero-division protection
  - [ ] 4.10 Create `internal/api/mcp/controllers/math.go` MCP controller stub
  - [ ] 4.11 Write test for MCP math tool registration and parameter validation
  - [ ] 4.12 Implement MCP math tool controller with mcp-go framework
  - [ ] 4.13 Write test for math tool execution through MCP protocol
  - [ ] 4.14 Test integration with MCP server and verify parameter type safety

- [x] 5.0 Integrate MCP configuration with existing config system
  - [x] 5.1 Add MCP configuration structure to existing config types
  - [x] 5.2 Update `config/default.json` with MCP server settings (port, host)
  - [x] 5.3 Update `config/provide.go` to inject MCP configuration
  - [x] 5.4 Update `cmd/mcp/root.go` to register MCP dependencies

- [ ] 6.0 Add comprehensive testing and documentation
  - [ ] 6.1 Create end-to-end test script for stdio transport
  - [ ] 6.2 Create end-to-end test script for HTTP transport
  - [ ] 6.3 Create usage examples and integration guide
  - [ ] 6.4 Document MCP tool development patterns
  - [ ] 6.5 Verify all tests pass and coverage meets requirements 