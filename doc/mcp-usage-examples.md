# MCP Usage Examples and Integration Guide

This document provides comprehensive examples and integration guidance for using the MCP (Model Context Protocol) tools extension in the golang-backend-boilerplate project.

## Overview

The MCP tools extension provides mathematical and time-related operations through both stdio and HTTP transports. It includes:

- **Time Tool**: Get current time in various formats
- **Math Tools**: Basic arithmetic operations (add, subtract, multiply, divide, calculate)

## Quick Start

### 1. Build the MCP Server

```bash
go build -o bin/mcp cmd/mcp/main.go
```

### 2. Start with stdio Transport

```bash
./bin/mcp stdio
```

### 3. Start with HTTP Transport

```bash
./bin/mcp http --host localhost --port 8080
```

## Transport Options

### stdio Transport

Best for:
- Direct integration with applications
- Piping data through command line
- Lightweight communication

**Example:**
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/mcp stdio
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

## Available Tools

### Time Tool: `current_time`

Get the current time in various formats.

**Parameters:**
- `format` (string): Time format - "iso", "unix", or "rfc3339"

**Examples:**

```json
// Get current time in ISO format
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "current_time",
    "arguments": {
      "format": "iso"
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Current time (iso): 2024-01-15T10:30:45Z"
      }
    ]
  }
}
```

### Math Tools

#### Addition: `add`

**Parameters:**
- `a` (number): First number
- `b` (number): Second number

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "add",
    "arguments": {
      "a": 5,
      "b": 3
    }
  }
}
```

#### Subtraction: `subtract`

**Parameters:**
- `a` (number): Number to subtract from
- `b` (number): Number to subtract

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "subtract",
    "arguments": {
      "a": 10,
      "b": 4
    }
  }
}
```

#### Multiplication: `multiply`

**Parameters:**
- `a` (number): First number
- `b` (number): Second number

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "multiply",
    "arguments": {
      "a": 6,
      "b": 7
    }
  }
}
```

#### Division: `divide`

**Parameters:**
- `a` (number): Dividend (number to be divided)
- `b` (number): Divisor (number to divide by)

**Note:** Division by zero will return an error.

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "divide",
    "arguments": {
      "a": 15,
      "b": 3
    }
  }
}
```

#### Generic Calculator: `calculate`

**Parameters:**
- `operation` (string): Operation type - "add", "subtract", "multiply", or "divide"
- `a` (number): First number
- `b` (number): Second number

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "calculate",
    "arguments": {
      "operation": "multiply",
      "a": 4,
      "b": 5
    }
  }
}
```

## Integration Examples

### Go Application Integration

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type MCPRequest struct {
    JsonRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
}

type MCPResponse struct {
    JsonRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Result  interface{} `json:"result"`
    Error   interface{} `json:"error,omitempty"`
}

func callMCPTool(baseURL, toolName string, args interface{}) (*MCPResponse, error) {
    request := MCPRequest{
        JsonRPC: "2.0",
        ID:      1,
        Method:  "tools/call",
        Params: map[string]interface{}{
            "name":      toolName,
            "arguments": args,
        },
    }

    jsonData, err := json.Marshal(request)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(baseURL+"/mcp", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var mcpResp MCPResponse
    if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
        return nil, err
    }

    return &mcpResp, nil
}

func main() {
    baseURL := "http://localhost:8080"
    
    // Add two numbers
    result, err := callMCPTool(baseURL, "add", map[string]interface{}{
        "a": 10,
        "b": 5,
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Addition result: %+v\n", result.Result)
    
    // Get current time
    timeResult, err := callMCPTool(baseURL, "current_time", map[string]interface{}{
        "format": "iso",
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Current time: %+v\n", timeResult.Result)
}
```

### Python Integration

```python
import json
import requests

class MCPClient:
    def __init__(self, base_url):
        self.base_url = base_url
        
    def call_tool(self, tool_name, arguments):
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        response = requests.post(
            f"{self.base_url}/mcp",
            json=request,
            headers={"Content-Type": "application/json"}
        )
        
        return response.json()
    
    def list_tools(self):
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/list"
        }
        
        response = requests.post(
            f"{self.base_url}/mcp",
            json=request,
            headers={"Content-Type": "application/json"}
        )
        
        return response.json()

# Usage
client = MCPClient("http://localhost:8080")

# List available tools
tools = client.list_tools()
print("Available tools:", tools)

# Perform calculations
add_result = client.call_tool("add", {"a": 15, "b": 25})
print("15 + 25 =", add_result)

multiply_result = client.call_tool("multiply", {"a": 6, "b": 7})
print("6 × 7 =", multiply_result)

# Get current time
time_result = client.call_tool("current_time", {"format": "unix"})
print("Current Unix time:", time_result)
```

### JavaScript/Node.js Integration

```javascript
const axios = require('axios');

class MCPClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
    }

    async callTool(toolName, arguments) {
        const request = {
            jsonrpc: "2.0",
            id: 1,
            method: "tools/call",
            params: {
                name: toolName,
                arguments: arguments
            }
        };

        try {
            const response = await axios.post(`${this.baseURL}/mcp`, request, {
                headers: {
                    'Content-Type': 'application/json'
                }
            });
            return response.data;
        } catch (error) {
            throw new Error(`MCP call failed: ${error.message}`);
        }
    }

    async listTools() {
        const request = {
            jsonrpc: "2.0",
            id: 1,
            method: "tools/list"
        };

        const response = await axios.post(`${this.baseURL}/mcp`, request);
        return response.data;
    }
}

// Usage
async function main() {
    const client = new MCPClient('http://localhost:8080');

    try {
        // List tools
        const tools = await client.listTools();
        console.log('Available tools:', tools);

        // Perform math operations
        const addResult = await client.callTool('add', { a: 20, b: 30 });
        console.log('20 + 30 =', addResult);

        const divideResult = await client.callTool('divide', { a: 100, b: 4 });
        console.log('100 ÷ 4 =', divideResult);

        // Get current time
        const timeResult = await client.callTool('current_time', { format: 'rfc3339' });
        console.log('Current time (RFC3339):', timeResult);

    } catch (error) {
        console.error('Error:', error.message);
    }
}

main();
```

## Error Handling

### Common Error Scenarios

1. **Division by Zero**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Division failed: division by zero: cannot divide 10.000000 by 0.000000"
      }
    ],
    "isError": true
  }
}
```

2. **Invalid Tool Name**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Tool not found: nonexistent_tool"
  }
}
```

3. **Invalid Parameters**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Invalid parameters: a parameter is required"
      }
    ],
    "isError": true
  }
}
```

## Testing

### Run End-to-End Tests

**stdio Transport:**
```bash
./scripts/test-mcp-stdio.sh
```

**HTTP Transport:**
```bash
./scripts/test-mcp-http.sh
```

### Manual Testing

**Test Tool Discovery:**
```bash
# stdio
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/mcp stdio

# HTTP
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

**Test Math Operations:**
```bash
# stdio - Addition
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add","arguments":{"a":5,"b":3}}}' | ./bin/mcp stdio

# HTTP - Multiplication
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"multiply","arguments":{"a":6,"b":7}}}'
```

## Configuration

### Default Configuration

The MCP server uses the existing configuration system. Default settings can be found in `config/default.json`:

```json
{
  "mcp": {
    "http": {
      "host": "localhost",
      "port": 8080
    }
  }
}
```

### Environment Variables

Override configuration with environment variables:

```bash
export MCP_HTTP_HOST=0.0.0.0
export MCP_HTTP_PORT=9090
./bin/mcp http
```

## Performance Considerations

### stdio Transport
- **Pros**: Lower latency, direct process communication
- **Cons**: Limited to single concurrent connection
- **Best for**: Command-line tools, single-user applications

### HTTP Transport
- **Pros**: Multiple concurrent connections, web-friendly
- **Cons**: HTTP overhead, network latency
- **Best for**: Web services, distributed systems

### Scaling Recommendations

1. **Load Balancing**: Use multiple HTTP server instances behind a load balancer
2. **Caching**: Cache frequently computed results at the application level
3. **Connection Pooling**: Reuse HTTP connections for better performance
4. **Monitoring**: Monitor response times and error rates

## Security Considerations

1. **Input Validation**: All parameters are validated for type and presence
2. **Error Handling**: Errors don't expose internal system details
3. **Access Control**: Consider adding authentication for HTTP transport
4. **Rate Limiting**: Implement rate limiting for production deployments

## Troubleshooting

### Common Issues

1. **Server Won't Start**
   - Check if port is already in use
   - Verify configuration file syntax
   - Check logs for detailed error messages

2. **Tool Not Found**
   - Verify tool name spelling
   - Check if server started successfully
   - List available tools using `tools/list` method

3. **Parameter Errors**
   - Ensure all required parameters are provided
   - Check parameter types (numbers should be numeric, not strings)
   - Validate parameter ranges for division operations

### Debug Mode

Enable debug logging by setting the log level:

```bash
export LOG_LEVEL=debug
./bin/mcp stdio
```

This provides detailed information about:
- Tool registration
- Request processing
- Parameter validation
- Response generation

## Contributing

To add new MCP tools:

1. Create service in `internal/app/`
2. Add corresponding tests
3. Create MCP controller in `internal/api/mcp/controllers/`
4. Register in `internal/api/mcp/controllers/register.go`
5. Update documentation and tests

For detailed development patterns, see `doc/mcp-development-patterns.md`. 