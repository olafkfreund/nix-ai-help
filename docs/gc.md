# nixai gc

AI-guided intelligent garbage collection with safe cleanup, generation management, and storage optimization.

---

## 🆕 TUI Integration & Enhanced Features

The `nixai gc` command now features **comprehensive TUI integration** with intelligent cleanup capabilities:

### ✨ **Interactive TUI Features**
- **🎯 Interactive Parameter Input**: Cleanup options and safety settings through modern TUI interface
- **📊 Real-Time Cleanup Analysis**: Live space analysis and cleanup progress within the TUI
- **⌨️ Command Discovery**: Enhanced command browser with `[INPUT]` indicators for all subcommands and flags
- **🛡️ Interactive Safety Verification**: AI-powered safety checks with confirmation prompts
- **📋 Context-Aware Cleanup**: Automatic analysis of system usage patterns for safe cleanup

### 🗑️ **AI-Guided Garbage Collection Features**
- **🧠 Intelligent Safety Analysis**: AI-powered analysis to prevent accidental deletion of critical paths
- **📊 Storage Usage Analytics**: Detailed breakdown of store usage with optimization recommendations
- **🎯 Smart Generation Management**: Intelligent selection of generations to keep based on usage patterns
- **🔒 Rollback Protection**: Automatic protection of generations needed for system recovery
- **📈 Cleanup Impact Prediction**: Estimate space savings and potential risks before cleanup
- **⚡ Performance-Aware Cleanup**: Optimize cleanup operations for minimal system impact
- **🔍 Dependency Analysis**: Deep analysis of store path dependencies to prevent broken references

## Command Structure

```sh
nixai gc [subcommand] [flags]
```

### Available Subcommands (4 Total)

| Subcommand | Description | TUI Support |
|------------|-------------|-------------|
| `analyze` | Analyze storage usage and recommend cleanup actions | ✅ Interactive |
| `generations` | Manage system and user generations with AI guidance | ✅ Interactive |
| `store-paths` | Clean specific store paths with safety verification | ✅ Interactive |
| `optimize` | Optimize store with deduplication and compression | ✅ Interactive |

### Enhanced Flags (3 Total)

| Flag | Description | TUI Input |
|------|-------------|-----------|
| `--dry-run` | Show what would be deleted without actually deleting | ✅ Interactive |
| `--older-than <duration>` | Only delete generations older than given duration | ✅ Interactive |
| `--aggressive` | More aggressive cleanup with AI safety validation | ✅ Interactive |

## Command Help Output

```sh
./nixai gc --help
Run garbage collection and clean up unused Nix store paths.

Usage:
  nixai gc [flags]

Flags:
  -h, --help   help for gc
  --dry-run    Show what would be deleted without actually deleting
  --older-than duration   Only delete generations older than the given duration (e.g. 30d)

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai gc
  nixai gc --dry-run
  nixai gc --older-than 30d
```

---

## Real Life Examples

- **Free up disk space after a big update:**
  ```sh
  nixai gc
  # Cleans up unused store paths and generations
  ```
- **Preview what will be deleted:**
  ```sh
  nixai gc --dry-run
  # Shows a list of items that would be removed
  ```
- **Remove only generations older than 30 days:**
  ```sh
  nixai gc --older-than 30d
  # Keeps recent generations, deletes older ones
  ```
