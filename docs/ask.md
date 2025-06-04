# nixai ask

Ask any NixOS-related question and get an AI-powered answer.

---

## Command Help Output

```sh
./nixai ask --help
Ask any NixOS-related question and get an AI-powered answer.

Usage:
  nixai "your question here"
  nixai --ask "your question here"

Flags:
  -h, --help   help for ask

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai "How do I enable Bluetooth on NixOS?"
  nixai --ask "What is a Nix flake?"
```

---

## Real Life Examples

- **Ask about enabling a service:**
  ```sh
  nixai "How do I enable SSH in NixOS?"
  ```
- **Ask about troubleshooting:**
  ```sh
  nixai --ask "Why is my system not booting after an update?"
  ```
