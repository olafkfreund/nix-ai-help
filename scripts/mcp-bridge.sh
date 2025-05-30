#!/bin/bash
# VS Code MCP Bridge Script for nixai
# This script connects to the nixai MCP server via Unix socket

exec socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock
