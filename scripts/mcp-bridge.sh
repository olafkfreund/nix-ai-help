#!/bin/bash
# VS Code MCP Bridge Script for nixai
# Enhanced bridge script with context-aware features and error handling

set -euo pipefail

# Configuration
SOCKET_PATH="${NIXAI_MCP_SOCKET:-/tmp/nixai-mcp.sock}"
MAX_RETRIES=3
RETRY_DELAY=1

# Function to check if MCP server is running
check_server() {
    if [[ ! -S "$SOCKET_PATH" ]]; then
        echo "Error: MCP server socket not found at $SOCKET_PATH" >&2
        echo "Try starting the server with: nixai mcp-server start -d" >&2
        return 1
    fi
}

# Function to connect with retries
connect_with_retry() {
    local attempt=1
    
    while [[ $attempt -le $MAX_RETRIES ]]; do
        if check_server; then
            # Successfully connected
            exec socat STDIO UNIX-CONNECT:"$SOCKET_PATH"
            return 0
        fi
        
        echo "Attempt $attempt/$MAX_RETRIES failed, retrying in ${RETRY_DELAY}s..." >&2
        sleep $RETRY_DELAY
        ((attempt++))
    done
    
    echo "Failed to connect after $MAX_RETRIES attempts" >&2
    exit 1
}

# Main execution
connect_with_retry
