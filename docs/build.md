# nixai build

Build NixOS configurations, packages, or flakes using AI assistance.

---

## Command Help Output

```sh
./nixai build --help
Build NixOS configurations, packages, or flakes using AI assistance.

Usage:
  nixai build [target] [flags]

Flags:
  -h, --help   help for build
  --flake      Build a flake-based configuration

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai build system
  nixai build .#my-machine
  nixai build --flake
```

---

## Real Life Examples

- **Build the current system configuration:**
  ```sh
  nixai build system
  ```
- **Build a specific flake target:**
  ```sh
  nixai build .#my-machine
  ```
- **Build with flake support:**
  ```sh
  nixai build --flake
  ```
