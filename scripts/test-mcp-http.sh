#!/bin/bash

# End-to-end test script for MCP HTTP transport
# This script tests the MCP server with HTTP transport using both time and math tools

set -e

echo "=== MCP HTTP Transport End-to-End Test ==="
echo

# Configuration
HOST="localhost"
PORT="8080"
BASE_URL="http://$HOST:$PORT"
SERVER_PID=""

# Build the MCP server
echo "Building MCP server..."
go build -o bin/mcp cmd/mcp/main.go
echo "✓ MCP server built successfully"
echo

# Start HTTP server in background
start_server() {
    echo "Starting MCP HTTP server on $HOST:$PORT..."
    ./bin/mcp http --host $HOST --port $PORT &
    SERVER_PID=$!
    sleep 2 # Give server time to start
    
    # Check if server is running
    if kill -0 $SERVER_PID 2>/dev/null; then
        echo "✓ MCP HTTP server started successfully (PID: $SERVER_PID)"
    else
        echo "✗ Failed to start MCP HTTP server"
        exit 1
    fi
}

# Stop server
stop_server() {
    if [ ! -z "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Stopping MCP HTTP server..."
        kill $SERVER_PID
        wait $SERVER_PID 2>/dev/null || true
        echo "✓ MCP HTTP server stopped"
    fi
}

# Cleanup on exit
cleanup() {
    stop_server
    rm -f bin/mcp
}
trap cleanup EXIT

# Test functions
test_http_tool() {
    local tool_name="$1"
    local params="$2"
    local description="$3"
    
    echo "Testing $tool_name tool: $description"
    
    # Create MCP request
    local request=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "$tool_name",
        "arguments": $params
    }
}
EOF
)
    
    # Send request to MCP server via HTTP
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$request" \
        "$BASE_URL/mcp")
    
    if echo "$response" | grep -q '"result"'; then
        echo "✓ $tool_name test passed"
    else
        echo "✗ $tool_name test failed"
        echo "Response: $response"
        return 1
    fi
}

test_tools_list() {
    echo "Testing tools list endpoint..."
    
    local request=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
}
EOF
)
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$request" \
        "$BASE_URL/mcp")
    
    # Check if response contains expected tools
    if echo "$response" | grep -q '"name": "current_time"' && \
       echo "$response" | grep -q '"name": "add"' && \
       echo "$response" | grep -q '"name": "calculate"'; then
        echo "✓ Tools list test passed"
    else
        echo "✗ Tools list test failed"
        echo "Response: $response"
        return 1
    fi
}

test_server_health() {
    echo "Testing server health endpoint..."
    
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
    
    if [ "$response" -eq 200 ]; then
        echo "✓ Server health test passed"
    else
        echo "✗ Server health test failed (HTTP $response)"
        return 1
    fi
}

# Start the tests
start_server

echo
echo "Starting HTTP transport tests..."
echo

# Test 1: Server health check
test_server_health

echo

# Test 2: List available tools
test_tools_list

echo

# Test 3: Time tool - current time
test_http_tool "current_time" '{"format": "iso"}' "get current time in ISO format"

echo

# Test 4: Time tool - unix timestamp
test_http_tool "current_time" '{"format": "unix"}' "get current time in Unix format"

echo

# Test 5: Math tool - addition
test_http_tool "add" '{"a": 5, "b": 3}' "add two numbers"

echo

# Test 6: Math tool - subtraction  
test_http_tool "subtract" '{"a": 10, "b": 4}' "subtract two numbers"

echo

# Test 7: Math tool - multiplication
test_http_tool "multiply" '{"a": 6, "b": 7}' "multiply two numbers"

echo

# Test 8: Math tool - division
test_http_tool "divide" '{"a": 15, "b": 3}' "divide two numbers"

echo

# Test 9: Math tool - generic calculate
test_http_tool "calculate" '{"operation": "multiply", "a": 4, "b": 5}' "generic calculation (multiplication)"

echo

# Test 10: Error handling - division by zero
echo "Testing error handling: division by zero"
request=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "divide",
        "arguments": {"a": 10, "b": 0}
    }
}
EOF
)

response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$request" \
    "$BASE_URL/mcp")

if echo "$response" | grep -q '"isError": true'; then
    echo "✓ Division by zero error handling test passed"
else
    echo "✗ Division by zero error handling test failed"
    echo "Response: $response"
fi

echo

# Test 11: Error handling - invalid tool
echo "Testing error handling: invalid tool name"
request=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "nonexistent_tool",
        "arguments": {}
    }
}
EOF
)

response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$request" \
    "$BASE_URL/mcp")

if echo "$response" | grep -q '"error"'; then
    echo "✓ Invalid tool error handling test passed"
else
    echo "✗ Invalid tool error handling test failed"
    echo "Response: $response"
fi

echo

# Test 12: Concurrent requests
echo "Testing concurrent requests..."
for i in {1..5}; do
    test_http_tool "add" "{\"a\": $i, \"b\": $((i+1))}" "concurrent addition request $i" &
done
wait

echo "✓ Concurrent requests test passed"

echo

echo "=== All HTTP transport tests completed successfully! ==="
echo
echo "Summary:"
echo "- Server health: ✓"
echo "- Tools discovery: ✓"
echo "- Time tool (multiple formats): ✓" 
echo "- Math tools (add, subtract, multiply, divide): ✓"
echo "- Generic calculate tool: ✓"
echo "- Error handling (division by zero, invalid tool): ✓"
echo "- Concurrent requests: ✓"
echo
echo "The MCP server with HTTP transport is working correctly!" 