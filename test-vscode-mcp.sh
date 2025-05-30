#!/bin/bash

echo "🚀 VS Code MCP Integration Test"
echo "================================"

echo "1. 🔌 Testing socket connection..."
if [ -S "/tmp/nixai-mcp.sock" ]; then
    echo "✅ Unix socket exists"
else
    echo "❌ Unix socket missing"
    exit 1
fi

echo ""
echo "2. ⚙️ Checking VS Code configuration..."
if [ -f ".vscode/settings.json" ]; then
    echo "✅ Workspace settings.json exists"
else
    echo "❌ Workspace settings.json missing"
fi

if [ -f ".vscode/mcp-settings.json" ]; then
    echo "✅ Workspace mcp-settings.json exists"
else
    echo "❌ Workspace mcp-settings.json missing"
fi

echo ""
echo "3. 🧩 Checking VS Code extensions..."
MCP_EXTENSIONS=(
    "automatalabs.copilot-mcp"
    "saoudrizwan.claude-dev"
    "zebradev.mcp-server-runner"
)

for ext in "${MCP_EXTENSIONS[@]}"; do
    if code --list-extensions 2>/dev/null | grep -q "$ext"; then
        echo "✅ Extension installed: $ext"
    else
        echo "❌ Extension missing: $ext"
    fi
done

echo ""
echo "4. 🧪 Testing MCP protocol..."
timeout 5 socat - UNIX-CONNECT:/tmp/nixai-mcp.sock <<< '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | head -1 | grep -q "result"

if [ $? -eq 0 ]; then
    echo "✅ MCP protocol working"
else
    echo "❌ MCP protocol test failed"
fi

echo ""
echo "🎯 Next Steps:"
echo "1. Open VS Code: code ."
echo "2. Reload window: Ctrl+Shift+P -> 'Developer: Reload Window'"
echo "3. Check for 'nixai' MCP server in extensions"
echo "4. Test in Copilot Chat or Claude extensions"
