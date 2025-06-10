# nixai hardware

Comprehensive hardware detection, optimization, and management with AI-powered analysis and automated configuration generation.

Analyze your system hardware and get AI-powered recommendations for optimal NixOS configuration including drivers, performance settings, power management, and hardware-specific optimizations.

---

## 🆕 TUI Integration & Enhanced Features

The `nixai hardware` command now features **comprehensive TUI integration** with advanced hardware management:

### ✨ **Interactive TUI Features**
- **🎯 Interactive Parameter Input**: Hardware analysis options and optimization settings through modern TUI interface
- **📊 Real-Time Hardware Analysis**: Live hardware detection and analysis progress within the TUI
- **⌨️ Command Discovery**: Enhanced command browser with `[INPUT]` indicators for all 6 subcommands and 6 flags
- **🔧 Interactive Optimization Wizard**: Step-by-step hardware optimization with AI guidance
- **📋 Context-Aware Hardware Detection**: Integration with existing NixOS configuration for seamless optimization

### 💻 **Advanced Hardware Management Features**
- **🧠 AI-Powered Hardware Analysis**: Machine learning-based hardware detection with optimization recommendations
- **⚡ Performance Optimization**: Automatic CPU, GPU, and memory optimization based on workload patterns
- **🔋 Power Management**: Intelligent power configuration for laptops and servers with usage-based tuning
- **🎮 Gaming Optimization**: Specialized configurations for gaming hardware with latency and performance tuning
- **🖥️ Multi-GPU Support**: Automatic detection and configuration of multiple GPUs with workload distribution
- **🌡️ Thermal Management**: Intelligent cooling configuration with temperature monitoring and alerts
- **🔌 Hardware Compatibility**: Comprehensive hardware compatibility checking with alternative suggestions

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
  function                - Use hardware function calling interface

Examples:
  nixai hardware detect
  nixai hardware optimize --dry-run
  nixai hardware drivers --auto-install
  nixai hardware laptop --power-save
  nixai hardware function --operation detect

Usage:
  nixai hardware [flags]
  nixai hardware [command]

Available Commands:
  compare     Compare current vs optimal settings
  detect      Detect and analyze system hardware
  drivers     Auto-configure drivers and firmware
  function    Use hardware function calling interface
  laptop      Laptop-specific optimizations
  optimize    Apply hardware-specific optimizations

Flags:
  -h, --help   help for hardware

Global Flags:
      --agent string          Specify the agent type (ollama, openai, gemini, etc.)
  -a, --ask string            Ask a question about NixOS configuration
      --context-file string   Path to a file containing context information (JSON or text)
  -n, --nixos-path string     Path to your NixOS configuration folder (containing flake.nix or configuration.nix)
      --role string           Specify the agent role (diagnoser, explainer, ask, build, etc.)
      --tui                   Launch TUI mode for any command

Use "nixai hardware [command] --help" for more information about a command.
```

---

## Latest Features & Capabilities

### 🆕 New Hardware Function Interface
- **Structured Operations**: Use `nixai hardware function` for standardized hardware operations
- **Component-Specific Actions**: Target specific hardware components (GPU, CPU, storage, network)
- **Multiple Output Formats**: Generate configurations in Nix, JSON, or text formats
- **Advanced Diagnostics**: Detailed hardware issue detection and resolution

### 🔧 Enhanced Driver Management
- **Auto-Installation Support**: Get ready-to-run installation commands with `--auto-install`
- **Comprehensive Driver Database**: Support for GPU, WiFi, Bluetooth, audio, and peripheral drivers
- **Firmware Management**: Automatic detection and configuration of hardware firmware
- **Microcode Updates**: CPU microcode optimization recommendations

### ⚡ Intelligent Optimization Engine
- **AI-Powered Recommendations**: Context-aware optimization suggestions
- **Dry-Run Mode**: Preview changes before applying with `--dry-run`
- **Performance vs Power Balance**: Smart recommendations based on hardware type
- **Storage-Specific Tuning**: SSD TRIM, NVMe optimization, and I/O scheduler selection

### 💻 Advanced Laptop Support
- **Dual Performance Modes**: Choose between `--power-save` and `--performance`
- **Thermal Management**: Advanced CPU and GPU thermal control
- **Docking Station Support**: External monitor and peripheral configuration
- **Hibernation Optimization**: Smart suspend and hibernation setup

### 🔍 Comprehensive Hardware Detection
- **Multi-Component Analysis**: CPU features, GPU capabilities, memory configuration
- **Network Interface Optimization**: Ethernet and WiFi performance tuning
- **Storage Configuration**: RAID, encryption, and filesystem recommendations
- **Virtualization Support**: Hardware virtualization feature detection

### 🎯 Integration Features
- **AI Agent Integration**: Works with multiple AI providers (Ollama, OpenAI, Gemini)
- **TUI Mode Support**: Interactive hardware management with `--tui`
- **Context-Aware**: Uses `--context-file` for enhanced recommendations
- **NixOS Path Integration**: Automatic configuration file detection with `--nixos-path`

### 📊 Configuration Management
- **Drift Detection**: Compare current vs optimal configurations
- **Best Practice Compliance**: Ensure configurations follow NixOS best practices
- **Version Compatibility**: Check hardware compatibility across NixOS versions
- **Migration Assistance**: Help migrate configurations between systems

---

## Real Life Examples

### 🔍 Hardware Detection and Analysis

- **Detect all hardware components:**

  ```sh
  nixai hardware detect
  ```

  Output includes:
  - CPU model and features (e.g., AMD Ryzen Threadripper PRO 3995WX 64-Cores)
  - GPU information (e.g., AMD Radeon RX 7900 XT/XTX)
  - Memory configuration (e.g., Total RAM: 219Gi)
  - Storage devices and configuration
  - Network interfaces and status

### ⚡ Hardware Optimization

- **Preview optimization recommendations:**

  ```sh
  nixai hardware optimize --dry-run
  ```

  Shows recommendations for:
  - CPU-specific optimization flags
  - Memory management settings
  - Storage optimization (SSD TRIM, I/O schedulers)
  - GPU acceleration settings
  - Network interface tuning

- **Apply optimizations with AI guidance:**

  ```sh
  nixai hardware optimize
  ```

### 🔧 Driver Management

- **Auto-detect and configure drivers:**

  ```sh
  nixai hardware drivers
  ```

- **Get installation commands for missing drivers:**

  ```sh
  nixai hardware drivers --auto-install
  ```

  Handles:
  - GPU drivers (NVIDIA, AMD, Intel)
  - WiFi and Bluetooth firmware
  - Audio drivers and codec configuration
  - USB and peripheral drivers

### 💻 Laptop Optimizations

- **Optimize for maximum battery life:**

  ```sh
  nixai hardware laptop --power-save
  ```

- **Optimize for maximum performance:**

  ```sh
  nixai hardware laptop --performance
  ```

  Configures:
  - Power management and thermal control
  - Display scaling and brightness
  - WiFi power saving
  - Suspend/hibernation setup

### 📊 Configuration Comparison

- **Compare current vs optimal settings:**

  ```sh
  nixai hardware compare
  ```

  Analyzes:
  - Current driver configurations vs recommended
  - Missing optimization opportunities
  - Potential compatibility issues
  - Best practice compliance

### 🔬 Advanced Function Interface

- **Use structured hardware operations:**

  ```sh
  nixai hardware function --operation detect
  ```

- **Generate NixOS configuration for specific components:**

  ```sh
  nixai hardware function --operation generate-config --component gpu --format nix
  ```

- **Perform comprehensive hardware scan:**

  ```sh
  nixai hardware function --operation scan --detailed
  ```

- **Diagnose hardware issues:**

  ```sh
  nixai hardware function --operation diagnose --include-drivers
  ```

### 🎯 Real-World Scenarios

- **New laptop setup with WiFi issues:**

  ```sh
  # Detect hardware and missing drivers
  nixai hardware detect
  
  # Configure WiFi drivers automatically
  nixai hardware drivers --auto-install
  
  # Optimize for laptop usage
  nixai hardware laptop --power-save
  ```

- **Gaming workstation optimization:**

  ```sh
  # Detect high-end hardware
  nixai hardware detect
  
  # Optimize for performance
  nixai hardware optimize
  
  # Configure GPU drivers and acceleration
  nixai hardware drivers
  ```

- **Server hardware audit:**

  ```sh
  # Compare current config vs optimal
  nixai hardware compare
  
  # Generate optimized configuration
  nixai hardware function --operation generate-config --format nix
  ```

### 🚀 Integration with AI Agents

- **Ask hardware-specific questions:**

  ```sh
  nixai --ask "How do I configure AMD GPU acceleration for video encoding?"
  ```

- **Use TUI mode for interactive hardware management:**

  ```sh
  nixai hardware --tui
  ```

- **Specify agent type for hardware queries:**

  ```sh
  nixai hardware detect --agent ollama --role hardware-expert
  ```
