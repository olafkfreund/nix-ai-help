# nixai diagnose

Diagnose NixOS issues from logs, configs, or piped input using AI.

---

## Command Help Output

```sh
./nixai diagnose --help
Diagnose NixOS issues from logs, configs, or piped input using AI.

Usage:
  nixai diagnose [file|--pipe]

Flags:
  -h, --help   help for diagnose
  --pipe       Read input from stdin (for piped logs)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai diagnose /var/log/nixos.log
  cat /var/log/nixos.log | nixai diagnose --pipe
```

---

## Real Life Examples

- **Diagnose a failed system switch from a log file:**
  ```sh
  nixai diagnose /var/log/nixos.log
  # Analyzes the log and suggests fixes
  ```
- **Pipe journalctl output for diagnosis:**
  ```sh
  journalctl -xe | nixai diagnose --pipe
  # AI reviews the log and provides troubleshooting steps
  ```
