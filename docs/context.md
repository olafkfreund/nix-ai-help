# nixai context

Manage NixOS system context detection and caching for personalized, system-aware assistance.

---

## Overview

The `nixai context` command provides complete control over the intelligent context detection system that makes all nixai commands aware of your specific NixOS configuration. This system automatically detects and caches information about your system setup, including configuration type (flakes vs channels), Home Manager configuration, NixOS version, enabled services, and more.

## Context System Benefits

### 🎯 **Personalized AI Assistance**
Every nixai command now provides system-specific help tailored to your actual configuration:

```bash
# Example context-aware response
$ nixai ask "How do I enable SSH?"
📋 System: nixos | Flakes: Yes | Home Manager: standalone

✅ Since you're using flakes, add this to your flake.nix:
{
  services.openssh = {
    enable = true;
    settings.PasswordAuthentication = false;
  };
}

Then run: sudo nixos-rebuild switch --flake .
```

### ⚡ **Performance Optimized**
- Intelligent caching prevents repeated system scans
- Context detection happens once and is reused across all commands
- Automatic cache invalidation when configuration changes

### 🔍 **Comprehensive Detection**
The context system automatically detects:
- **Configuration Type**: Flakes vs traditional channels
- **Home Manager**: Standalone vs NixOS module integration
- **System Information**: NixOS version, Nix version, architecture
- **Services**: Currently enabled systemd services
- **Packages**: System and user-level installed packages
- **File Locations**: Configuration paths and flake files

---

## Command Help Output

```bash
$ nixai context --help
Manage the NixOS system context detection and caching system.

The context system automatically detects your NixOS configuration details including:
- Flakes vs channels usage
- Home Manager configuration
- NixOS version and system type
- Enabled services and installed packages
- Configuration file locations

This context information is used throughout nixai to provide more relevant
and targeted assistance.

Examples:
  nixai context detect    # Force re-detect system context
  nixai context show     # Display current context information
  nixai context reset    # Clear cached context and force refresh
  nixai context status   # Show context system status

Usage:
  nixai context [command]

Available Commands:
  detect      Force re-detection of NixOS system context
  reset       Clear cached context and force refresh
  show        Display current NixOS system context
  status      Show context detection system status

Flags:
  -h, --help   help for context

Global Flags:
      --agent string          Specify the agent type (ollama, openai, gemini, etc.)
  -a, --ask string            Ask a question about NixOS configuration
      --context-file string   Path to a file containing context information (JSON or text)
  -n, --nixos-path string     Path to your NixOS configuration folder (containing flake.nix or configuration.nix)
      --role string           Specify the agent role (diagnoser, explainer, ask, build, etc.)
      --tui                   Launch TUI mode for any command

Use "nixai context [command] --help" for more information about a command.
```

---

## Subcommands

### `nixai context detect`

Force re-detection of your NixOS system context, ignoring any cached information.

**Usage:**
```bash
nixai context detect [flags]
```

**Flags:**
- `--format json, -f json`: Output context in JSON format
- `--verbose, -v`: Show detailed detection process

**Examples:**
```bash
# Basic re-detection
nixai context detect

# Verbose output showing detection steps
nixai context detect --verbose

# JSON output for scripting
nixai context detect --format json
```

**Sample Output:**
```bash
$ nixai context detect --verbose

🔍 NixOS Context Detection

Starting context detection process...
Clearing context cache...
Re-detecting system context...

System Summary: System: nixos | Flakes: Yes | Home Manager: standalone

### System Information
System Type: nixos
NixOS Version: 25.11.20250607.3e3afe5 (Xantusia)
Nix Version: nix (Nix) 2.28.3

### Configuration
Uses Flakes: ✅ Yes
Uses Channels: ✅ Yes
Has Home Manager: ✅ Yes
Home Manager Type: standalone

### File Paths
NixOS Config: /etc/nixos
Flake File: /etc/nixos/flake.nix
Configuration.nix: /etc/nixos/configuration.nix

### Cache Information
Cache Valid: ✅ Yes
Last Detected: 2025-06-10 18:56:02

✅ Context detection completed
Cache location: /home/user/.config/nixai/config.yaml
```

### `nixai context show`

Display the current NixOS system context information.

**Usage:**
```bash
nixai context show [flags]
```

**Flags:**
- `--format json, -f json`: Output context in JSON format
- `--detailed, -d`: Show detailed context information including services and packages

**Examples:**
```bash
# Basic context display
nixai context show

# Detailed view with services and packages
nixai context show --detailed

# JSON output for automation
nixai context show --format json
```

**Sample Output:**
```bash
$ nixai context show --detailed

📋 NixOS System Context

System Summary: System: nixos | Flakes: Yes | Home Manager: standalone

### System Information
System Type: nixos
NixOS Version: 25.11.20250607.3e3afe5 (Xantusia)
Nix Version: nix (Nix) 2.28.3

### Configuration
Uses Flakes: ✅ Yes
Uses Channels: ✅ Yes
Has Home Manager: ✅ Yes
Home Manager Type: standalone

### File Paths
NixOS Config: /etc/nixos
Home Manager Config: /home/user/.config/home-manager
Flake File: /etc/nixos/flake.nix
Configuration.nix: /etc/nixos/configuration.nix
Hardware Config: /etc/nixos/hardware-configuration.nix

### Enabled Services (15 total)
  • NetworkManager
  • systemd-resolved
  • openssh
  • docker
  • tailscaled
  • pipewire
  • xserver
  • displayManager
  • desktopManager
  • bluetooth
  ... and 5 more

### Installed Packages (127 total)
  • firefox
  • git
  • neovim
  • docker
  • tailscale
  • python3
  • nodejs
  • go
  • rust
  • gcc
  ... and 117 more

### Cache Information
Cache Valid: ✅ Yes
Last Detected: 2025-06-10 18:56:02
Cache Age: 2m15s
```

### `nixai context reset`

Clear the cached NixOS system context and force a fresh detection.

**Usage:**
```bash
nixai context reset [flags]
```

**Flags:**
- `--confirm, -y`: Skip confirmation prompt

**Examples:**
```bash
# Interactive reset with confirmation
nixai context reset

# Skip confirmation prompt
nixai context reset --confirm
```

**Sample Output:**
```bash
$ nixai context reset

🔄 Reset NixOS Context

This will clear all cached context information and force re-detection.

Continue with context reset? (y/N): y

Clearing context cache...
Re-detecting system context...

✅ Context reset completed

📋 System: nixos | Flakes: Yes | Home Manager: standalone
```

### `nixai context status`

Show the status of the context detection system including cache validity, health, and any errors.

**Usage:**
```bash
nixai context status [flags]
```

**Flags:**
- `--format json, -f json`: Output status in JSON format

**Examples:**
```bash
# Basic status check
nixai context status

# JSON status for monitoring
nixai context status --format json
```

**Sample Output:**
```bash
$ nixai context status

📊 Context System Status

Cache Location: /home/user/.config/nixai/config.yaml
Has Context: ✅ Yes
Cache Valid: ✅ Yes
Last Detected: 2025-06-10 18:56:02
Cache Age: 8m7s

✅ Context system is healthy
📋 System: nixos | Flakes: Yes | Home Manager: standalone
```

---

## JSON Output Format

All context commands support JSON output for automation and scripting:

```bash
$ nixai context show --format json
{
  "systemType": "nixos",
  "nixosVersion": "25.11.20250607.3e3afe5 (Xantusia)",
  "nixVersion": "nix (Nix) 2.28.3",
  "usesFlakes": true,
  "usesChannels": true,
  "hasHomeManager": true,
  "homeManagerType": "standalone",
  "nixosConfigPath": "/etc/nixos",
  "homeManagerConfigPath": "/home/user/.config/home-manager",
  "flakeFile": "/etc/nixos/flake.nix",
  "configurationNix": "/etc/nixos/configuration.nix",
  "hardwareConfigNix": "/etc/nixos/hardware-configuration.nix",
  "enabledServices": ["NetworkManager", "systemd-resolved", "openssh", "docker"],
  "installedPackages": ["firefox", "git", "neovim", "docker"],
  "cacheValid": true,
  "lastDetected": "2025-06-10T18:56:02Z",
  "detectionErrors": []
}
```

---

## Integration with Other Commands

### Context-Aware Command Output

All nixai commands now display context information when available:

```bash
$ nixai ask "How do I configure nginx?"
📋 System: nixos | Flakes: Yes | Home Manager: standalone

# AI response is now tailored to your flake-based setup...
```

```bash
$ nixai hardware detect
📋 System: nixos | Flakes: Yes | Home Manager: standalone

🔄 Detecting hardware components...
# Hardware detection proceeds with system awareness...
```

### Context in TUI Mode

The modern TUI interface shows context status:

```bash
$ nixai interactive
# Context information is displayed in the status area
```

---

## Troubleshooting

### Context Detection Issues

**Problem: Context detection fails**
```bash
$ nixai context status
❌ Context detection failed: permission denied

$ nixai context detect --verbose
# Shows detailed error information
```

**Solution:**
1. Check file permissions for `/etc/nixos` and configuration files
2. Ensure you have read access to system configuration
3. Try running with appropriate permissions if needed

**Problem: Outdated context information**
```bash
# After major system changes, context might be stale
$ nixai context reset
```

**Problem: Context shows wrong information**
```bash
# Force fresh detection
$ nixai context detect

# Check for detection errors
$ nixai context status
```

### Performance Considerations

**Cache Management:**
- Context is cached for performance
- Cache automatically invalidates after significant time or system changes
- Manual refresh available via `nixai context reset`

**Detection Frequency:**
- First run: Full detection (1-3 seconds)
- Subsequent runs: Cached access (instant)
- Automatic refresh: When cache expires or system changes detected

---

## Automation and Scripting

### Monitoring Context Health

```bash
#!/usr/bin/env bash
# Monitor context system health

status=$(nixai context status --format json)
cache_valid=$(echo "$status" | jq -r '.cache_valid')

if [[ "$cache_valid" != "true" ]]; then
    echo "Context cache invalid, refreshing..."
    nixai context detect
fi
```

### Configuration Change Detection

```bash
#!/usr/bin/env bash
# Detect if system configuration changed

old_context=$(nixai context show --format json)
nixai context detect >/dev/null 2>&1
new_context=$(nixai context show --format json)

if [[ "$old_context" != "$new_context" ]]; then
    echo "System configuration changed!"
    # Trigger any necessary updates
fi
```

### Context Information Extraction

```bash
#!/usr/bin/env bash
# Extract specific context information

context=$(nixai context show --format json)

uses_flakes=$(echo "$context" | jq -r '.usesFlakes')
nixos_version=$(echo "$context" | jq -r '.nixosVersion')
home_manager_type=$(echo "$context" | jq -r '.homeManagerType')

echo "System uses flakes: $uses_flakes"
echo "NixOS version: $nixos_version"
echo "Home Manager type: $home_manager_type"
```

---

## Real Life Examples

### Daily Development Workflow

```bash
# Morning system check
nixai context status

# Work on configuration
vim /etc/nixos/flake.nix

# After changes, refresh context
nixai context reset

# Get context-aware help
nixai ask "How do I optimize this configuration?"
```

### System Migration

```bash
# Before migration
nixai context show --format json > old-context.json

# Perform migration
nixai migrate channels-to-flakes

# After migration, verify context changed
nixai context detect
nixai context show

# Compare contexts
diff <(cat old-context.json | jq .) <(nixai context show --format json | jq .)
```

### Multi-Machine Management

```bash
#!/usr/bin/env bash
# Deploy to multiple machines with context awareness

for machine in desktop laptop server; do
    echo "Deploying to $machine..."
    
    # Check context on target machine
    ssh $machine "nixai context status"
    
    # Deploy with context-aware assistance
    ssh $machine "nixai machines deploy $machine"
done
```

### Automated Monitoring

```bash
#!/usr/bin/env bash
# Cron job to monitor system context health

log_file="/var/log/nixai-context-monitor.log"

{
    echo "$(date): Checking context system health..."
    
    status=$(nixai context status --format json 2>&1)
    if echo "$status" | jq -e '.cache_valid == true' >/dev/null 2>&1; then
        echo "$(date): Context system healthy"
    else
        echo "$(date): Context system needs attention"
        nixai context detect
        echo "$(date): Context refreshed"
    fi
    
    echo "$(date): Check completed"
    echo "---"
} >> "$log_file"
```

---

## Best Practices

### Regular Maintenance

1. **Monitor Context Health**: Use `nixai context status` regularly
2. **Refresh After Changes**: Run `nixai context reset` after major system modifications
3. **Verify Context Accuracy**: Use `nixai context show --detailed` to ensure detection is correct

### Integration with Workflows

1. **CI/CD Integration**: Include context validation in deployment scripts
2. **Configuration Management**: Use context detection to validate system state
3. **Documentation**: Include context information in system documentation

### Performance Optimization

1. **Cache Management**: Let nixai handle cache automatically unless issues arise
2. **Selective Refresh**: Use `nixai context detect` only when needed
3. **JSON Output**: Use JSON format for programmatic access to avoid parsing text

---

## Command Comparison

| Command | Purpose | Output | Use Case |
|---------|---------|---------|----------|
| `detect` | Force fresh detection | Context information | After system changes |
| `show` | Display current context | Context information | View current state |
| `reset` | Clear cache and refresh | Success message + context | Fix cache issues |
| `status` | System health check | Status information | Monitor system health |

---

**For complete command documentation and advanced usage, see the [User Manual](MANUAL.md).**
