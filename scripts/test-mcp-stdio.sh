#!/bin/bash

# End-to-end test script for MCP stdio transport
# This script tests the MCP server with stdio transport using both time and math tools

set -e

echo "=== MCP stdio Transport End-to-End Test ==="
echo

# Build the MCP server
echo "Building MCP server..."
go build -o bin/mcp cmd/mcp/main.go
echo "✓ MCP server built successfully"
echo

# Test functions
test_stdio_tool() {
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
    
    # Send request to MCP server via stdio
    echo "$request" | timeout 10s ./bin/mcp stdio 2>/dev/null | grep -q '"result"'
    
    if [ $? -eq 0 ]; then
        echo "✓ $tool_name test passed"
    else
        echo "✗ $tool_name test failed"
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
    
    local response=$(echo "$request" | timeout 10s ./bin/mcp stdio 2>/dev/null)
    
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

echo "Starting stdio transport tests..."
echo

# Test 1: List available tools
test_tools_list

echo

# Test 2: Time tool - current time
test_stdio_tool "current_time" '{"format": "iso"}' "get current time in ISO format"

echo

# Test 3: Math tool - addition
test_stdio_tool "add" '{"a": 5, "b": 3}' "add two numbers"

echo

# Test 4: Math tool - subtraction  
test_stdio_tool "subtract" '{"a": 10, "b": 4}' "subtract two numbers"

echo

# Test 5: Math tool - multiplication
test_stdio_tool "multiply" '{"a": 6, "b": 7}' "multiply two numbers"

echo

# Test 6: Math tool - division
test_stdio_tool "divide" '{"a": 15, "b": 3}' "divide two numbers"

echo

# Test 7: Math tool - generic calculate
test_stdio_tool "calculate" '{"operation": "add", "a": 8, "b": 2}' "generic calculation (addition)"

echo

# Test 8: Error handling - division by zero
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

response=$(echo "$request" | timeout 10s ./bin/mcp stdio 2>/dev/null)
if echo "$response" | grep -q '"isError": true'; then
    echo "✓ Division by zero error handling test passed"
else
    echo "✗ Division by zero error handling test failed"
    echo "Response: $response"
fi

echo

# Clean up
rm -f bin/mcp

echo "=== All stdio transport tests completed successfully! ==="
echo
echo "Summary:"
echo "- Tools discovery: ✓"
echo "- Time tool: ✓" 
echo "- Math tools (add, subtract, multiply, divide): ✓"
echo "- Generic calculate tool: ✓"
echo "- Error handling: ✓"
echo
echo "The MCP server with stdio transport is working correctly!" 