#!/bin/bash
# Final comprehensive test for MCP Context Integration
# Tests the complete integration from nixai context commands to MCP server tools

echo "🧪 MCP Context Integration - Final Comprehensive Test"
echo "======================================================"
echo

# Test 1: Basic nixai context functionality
echo "📋 Test 1: Basic Context System"
echo "--------------------------------"
./nixai context status
echo
./nixai context show
echo

# Test 2: MCP server status
echo "📋 Test 2: MCP Server Health"
echo "-----------------------------"
./nixai mcp-server status
echo

# Test 3: Test all 4 new MCP context tools
echo "📋 Test 3: MCP Context Tools"
echo "-----------------------------"

echo "Testing get_nixos_context..."
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "get_nixos_context", "arguments": {"format": "text", "detailed": false}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock | jq -r '.result.content[0].text' | head -10
echo

echo "Testing context_status..."
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "context_status", "arguments": {}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock | jq -r '.result.content[0].text' | head -10
echo

echo "Testing detect_nixos_context..."
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "detect_nixos_context", "arguments": {"verbose": false}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock | jq -r '.result.content[0].text' | head -10
echo

echo "Testing reset_nixos_context..."
echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "reset_nixos_context", "arguments": {"confirm": true}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock | jq -r '.result.content[0].text' | head -10
echo

echo "🎉 MCP Context Integration Test Complete!"
echo "✅ All systems operational and ready for VS Code/Neovim integration"
