# nixai migrate

AI-powered migration assistant for NixOS and Home Manager configurations with automated conversion, backup, and rollback support.

---

## 🆕 TUI Integration & Enhanced Features

The `nixai migrate` command now features **comprehensive TUI integration** with intelligent migration capabilities:

### ✨ **Interactive TUI Features**
- **🎯 Interactive Parameter Input**: Migration options and target configuration through modern TUI interface
- **📊 Real-Time Migration Progress**: Live migration analysis and conversion progress within the TUI
- **⌨️ Command Discovery**: Enhanced command browser with `[INPUT]` indicators for 2 subcommands and 3 flags
- **🔄 Interactive Migration Wizard**: Step-by-step migration with AI-guided recommendations
- **📋 Context-Aware Migration**: Automatic analysis of existing configuration for intelligent conversion

### 🔄 **AI-Powered Migration Features**
- **🧠 Intelligent Configuration Analysis**: AI-powered analysis of existing configurations with conversion strategies
- **📦 Channel-to-Flakes Migration**: Automated migration from channels to flakes with dependency resolution
- **🔧 Configuration Format Conversion**: Convert between configuration.nix, Home Manager, and flake.nix formats
- **🛡️ Automatic Backup System**: Comprehensive backup creation with versioning and restoration capabilities
- **✅ Migration Validation**: Pre and post-migration validation with rollback on failure
- **📊 Compatibility Analysis**: Deep analysis of configuration compatibility across NixOS versions
- **🔍 Dependency Migration**: Intelligent handling of package dependencies and service configurations

## Command Structure

```sh
nixai migrate [subcommand] [flags]
```

### Available Subcommands (2 Total)

| Subcommand | Description | TUI Support |
|------------|-------------|-------------|
| `to-flakes` | Migrate from channels to flakes with AI guidance | ✅ Interactive |
| `format` | Convert between configuration formats (config.nix ↔ flake.nix) | ✅ Interactive |

### Enhanced Flags (3 Total)

| Flag | Description | TUI Input |
|------|-------------|-----------|
| `--from <file>` | Source configuration file for migration | ✅ Interactive |
| `--to <file>` | Destination configuration file path | ✅ Interactive |
| `--backup` | Create automatic backup before migration | ✅ Interactive |

## Command Help Output

```sh
./nixai migrate --help
Migrate NixOS or Home Manager configurations.

Usage:
  nixai migrate [flags]

Flags:
  -h, --help   help for migrate
  --from FILE   Source configuration file
  --to FILE     Destination configuration file

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai migrate --from old.nix --to new.nix
```

---

## Real Life Examples

- **Migrate a configuration from old.nix to new.nix:**
  ```sh
  nixai migrate --from old.nix --to new.nix
  # Converts and adapts configuration to the new format
  ```
- **Migrate a Home Manager config:**
  ```sh
  nixai migrate --from home-old.nix --to home-new.nix
  # Migrates Home Manager settings
  ```
