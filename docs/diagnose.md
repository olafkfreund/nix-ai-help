# nixai diagnose

Advanced NixOS diagnostics with AI-powered log analysis, multi-format support, and intelligent issue detection and resolution.

---

## 🆕 TUI Integration & Enhanced Features

The `nixai diagnose` command now features **comprehensive TUI integration** with advanced diagnostic capabilities:

### ✨ **Interactive TUI Features**
- **🎯 Interactive Parameter Input**: File path selection and diagnostic options through modern TUI interface
- **📊 Real-Time Analysis Display**: Live diagnostic progress with AI insights within the TUI
- **⌨️ Command Discovery**: Enhanced command browser with `[INPUT]` indicators for configurable options
- **🔍 Interactive Flag Configuration**: All 4 flags (`--pipe`, `--format`, `--output`, `--severity`) configurable via TUI
- **📋 Context-Aware Diagnostics**: Automatic NixOS setup detection for personalized issue analysis

### 🩺 **Advanced Diagnostic Features**
- **🧠 AI-Powered Issue Detection**: Machine learning-based pattern recognition for complex NixOS problems
- **📊 Multi-Format Log Analysis**: Support for systemd journals, build logs, kernel logs, and custom formats
- **🎯 Intelligent Issue Classification**: Automatic categorization by severity, component, and resolution complexity
- **🔧 Automated Fix Suggestions**: AI-generated configuration patches and command-line remedies
- **📈 Trend Analysis**: Historical pattern recognition for recurring issues and performance degradation
- **🔍 Deep System Inspection**: Integration with `nixos-doctor`, hardware detection, and service status
- **📝 Comprehensive Reporting**: Detailed diagnostic reports with executive summaries and technical details

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
