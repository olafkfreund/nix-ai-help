# nixai hardware

Detect hardware and automatically generate optimized NixOS configurations.

Analyze your system hardware and get AI-powered recommendations for optimal NixOS configuration including drivers, performance settings, and power management.

---

## Command Help Output

```sh
./nixai hardware --help
Detect hardware and automatically generate optimized NixOS configurations.

Analyze your system hardware and get AI-powered recommendations for optimal
NixOS configuration including drivers, performance settings, and power management.

Commands:
  detect                  - Detect and analyze system hardware
  optimize                - Apply hardware-specific optimizations
  drivers                 - Auto-configure drivers and firmware
  compare                 - Compare current vs optimal settings
  laptop                  - Laptop-specific optimizations

Examples:
  nixai hardware detect
  nixai hardware optimize --dry-run
  nixai hardware drivers --auto-install
  nixai hardware laptop --power-save

Usage:
  nixai hardware [flags]
  nixai hardware [command]

Available Commands:
  compare     Compare current vs optimal settings
  detect      Detect and analyze system hardware
  drivers     Auto-configure drivers and firmware
  laptop      Laptop-specific optimizations
  optimize    Apply hardware-specific optimizations

Flags:
  -h, --help   help for hardware

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Use "nixai hardware [command] --help" for more information about a command.
```

---

## Real Life Examples

- **Diagnose missing drivers on a new laptop:**
  ```sh
  nixai hardware detect
  # Output will list missing drivers and suggest configuration changes
  ```
- **Optimize for battery life on a laptop:**
  ```sh
  nixai hardware laptop --power-save
  # Applies recommended power management settings
  ```
- **Compare current vs optimal hardware config:**
  ```sh
  nixai hardware compare
  # Shows differences and suggestions for improvement
  ```
