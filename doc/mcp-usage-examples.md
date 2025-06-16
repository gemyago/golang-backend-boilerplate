# MCP Usage Examples and Integration Guide

This document provides examples and integration guidance for using the MCP (Model Context Protocol) tools extension in the golang-backend-boilerplate project.

## Overview

The MCP tools extension provides mathematical and time-related operations through both stdio and HTTP transports. It includes:

- **Time Tool**: Get current time in various formats
- **Math Tools**: Basic arithmetic operations (add, subtract, multiply, divide, calculate)

## Quick Start

### 1. Start with stdio Transport

```bash
go run ./cmd/mcp/ stdio
```

### 2. Start with HTTP Transport

```bash
go run ./cmd/mcp/ http --host localhost --port 8080
```

## Transport Options

### stdio Transport

Best for:
- Direct integration with applications
- Piping data through command line
- Lightweight communication

**Example:**
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run ./cmd/mcp/ stdio
```

### HTTP Transport

Best for:
- Web applications and services
- Remote access
- Load balancing and scaling

**Example:**
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```
## Contributing

To add new MCP tools:

1. Create service in `internal/app/`
2. Add corresponding tests
3. Create MCP controller in `internal/api/mcp/controllers/`
4. Register in `internal/api/mcp/controllers/register.go`
5. Update documentation and tests