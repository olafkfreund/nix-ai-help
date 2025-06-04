# nixai flake

Manage and interact with Nix flakes.

---

## Command Help Output

```sh
./nixai flake --help
Manage and interact with Nix flakes.

Usage:
  nixai flake [command]

Available Commands:
  info      Show flake information
  update    Update flake inputs
  check     Check flake for errors
  init      Initialize a new flake

Flags:
  -h, --help   help for flake

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai flake info
  nixai flake update
  nixai flake check
  nixai flake init
```

---

## Real Life Examples

- **Show information about the current flake:**
  ```sh
  nixai flake info
  # Displays flake metadata, inputs, and outputs
  ```
- **Update all flake inputs:**
  ```sh
  nixai flake update
  # Updates all flake inputs to their latest versions
  ```
- **Check the flake for errors:**
  ```sh
  nixai flake check
  # Runs a check to ensure the flake is valid
  ```
- **Initialize a new flake in the current directory:**
  ```sh
  nixai flake init
  # Creates a new flake.nix template
  ```
