# nixai logs

Analyze and summarize NixOS logs using AI.

---

## Command Help Output

```sh
./nixai logs --help
Analyze and summarize NixOS logs using AI.

Usage:
  nixai logs [file|--pipe]

Flags:
  -h, --help   help for logs
  --pipe       Read input from stdin (for piped logs)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai logs /var/log/nixos.log
  cat /var/log/nixos.log | nixai logs --pipe
```

---

## Real Life Examples

- **Summarize a log file for errors:**
  ```sh
  nixai logs /var/log/nixos.log
  # AI summarizes and highlights issues in the log
  ```
- **Pipe journalctl output for analysis:**
  ```sh
  journalctl -xe | nixai logs --pipe
  # AI reviews the log and provides a summary
  ```
