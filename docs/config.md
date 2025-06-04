# nixai config

Manage nixai configuration settings.

---

## Command Help Output

```sh
./nixai config --help
Manage nixai configuration settings.

Usage:
  nixai config [get|set|edit] [key] [value]

Available Commands:
  get     View current configuration
  set     Set a configuration value
  edit    Edit the configuration file in your editor

Flags:
  -h, --help   help for config

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai config get
  nixai config set ai.provider ollama
  nixai config edit
```

---

## Real Life Examples

- **Switch AI provider to Gemini:**
  ```sh
  nixai config set ai.provider gemini
  # Changes the default AI provider to Gemini
  ```
- **Edit the configuration file in your editor:**
  ```sh
  nixai config edit
  # Opens the YAML config in your default editor
  ```
- **View all current configuration values:**
  ```sh
  nixai config get
  # Prints all current config settings
  ```
