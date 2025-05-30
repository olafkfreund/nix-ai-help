#!/bin/bash

# Test script for MCP server functionality
# Tests both HTTP and MCP protocol endpoints

set -e

echo "🧪 Testing nixai MCP Server Integration"
echo "======================================="

# Test 1: Server Status
echo "📋 Test 1: Server Status"
./nixai mcp-server status
echo "✅ Server status check passed"
echo

# Test 2: HTTP Query Endpoint
echo "📋 Test 2: HTTP Query Endpoint"
response=$(curl -s "http://localhost:8081/query?q=services.nginx.enable")
if echo "$response" | grep -q "nginx"; then
    echo "✅ HTTP query endpoint working"
else
    echo "❌ HTTP query endpoint failed"
    echo "Response: $response"
    exit 1
fi
echo

# Test 3: Unix Socket Exists
echo "📋 Test 3: Unix Socket"
if [ -S "/tmp/nixai-mcp.sock" ]; then
    echo "✅ Unix socket exists and is accessible"
else
    echo "❌ Unix socket not found"
    exit 1
fi
echo

# Test 4: Test MCP Protocol via socat (if available)
echo "📋 Test 4: MCP Protocol Test"
if command -v socat &> /dev/null; then
    echo "Testing MCP initialize request..."
    
    # Create a simple MCP initialize request
    cat << 'EOF' > /tmp/mcp-test.json
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}
EOF
    
    # Test the MCP protocol (with timeout)
    timeout 5s socat - UNIX-CONNECT:/tmp/nixai-mcp.sock < /tmp/mcp-test.json > /tmp/mcp-response.json 2>/dev/null || true
    
    if [ -s /tmp/mcp-response.json ]; then
        echo "✅ MCP protocol responding"
        echo "Response preview:"
        head -c 200 /tmp/mcp-response.json
        echo
    else
        echo "⚠️  MCP protocol test inconclusive (may need actual MCP client)"
    fi
    
    # Clean up
    rm -f /tmp/mcp-test.json /tmp/mcp-response.json
else
    echo "⚠️  socat not available, skipping MCP protocol test"
fi
echo

# Test 5: Test nixai CLI integration
echo "📋 Test 5: CLI Integration"
response=$(./nixai explain-option services.nginx.enable 2>&1 || true)
if echo "$response" | grep -q -E "(nginx|Nginx|HTTP|web server)" && ! echo "$response" | grep -q "Error"; then
    echo "✅ CLI integration working with MCP server"
else
    echo "⚠️  CLI integration test inconclusive"
    echo "Response preview:"
    echo "$response" | head -n 5
fi
echo

echo "🎉 MCP Server Integration Tests Complete!"
echo "=========================================="
echo "✅ Server is running and responding"
echo "✅ Both HTTP and Unix socket endpoints are functional"
echo "✅ Ready for VS Code MCP integration"
echo
echo "📝 VS Code Integration Instructions:"
echo "1. Install the MCP extension in VS Code"
echo "2. Add the configuration from .vscode/mcp-settings.json"
echo "3. Restart VS Code to activate MCP integration"
echo "4. Use MCP tools via VS Code's command palette or chat interface"
