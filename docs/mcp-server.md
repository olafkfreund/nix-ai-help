# nixai mcp-server

Interact with the Model Context Protocol (MCP) server for documentation queries.

---

## Command Help Output

```sh
./nixai mcp-server --help
Interact with the Model Context Protocol (MCP) server for documentation queries.

Usage:
  nixai mcp-server [command]

Available Commands:
  query      Query documentation from MCP server
  status     Show MCP server status
  restart    Restart the MCP server

Flags:
  -h, --help   help for mcp-server

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai mcp-server query "services.nginx.enable"
  nixai mcp-server status
```

---

## Real Life Examples

- **Query documentation for a NixOS option:**
  ```sh
  nixai mcp-server query "services.nginx.enable"
  # Returns documentation and usage for the option
  ```
- **Check MCP server status:**
  ```sh
  nixai mcp-server status
  # Shows if the MCP server is running and available
  ```
- **Restart the MCP server:**
  ```sh
  nixai mcp-server restart
  # Restarts the documentation server
  ```
