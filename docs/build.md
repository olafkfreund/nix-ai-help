# nixai build

Enhanced build troubleshooting and optimization with AI-powered analysis, real-time monitoring, and intelligent recovery capabilities.

---

## Overview

The `nixai build` command provides comprehensive build assistance with advanced features including:

- **AI-Powered Failure Analysis**: Deep analysis of build failures with intelligent fixes
- **Real-Time Monitoring**: Watch builds progress with live AI insights
- **Background Build Management**: Run builds in background with status tracking
- **Intelligent Recovery System**: Automated retry with learned optimizations from failure history
- **Performance Profiling**: Analyze and optimize build performance with detailed metrics
- **Cache Analysis**: Optimize binary cache usage and identify miss patterns
- **Sandbox Debugging**: Resolve permission and environment issues

---

## Command Structure

```sh
nixai build [subcommand] [args] [flags]
```

### Available Subcommands

| Subcommand | Description |
|------------|-------------|
| `debug <package>` | Deep build failure analysis with pattern recognition |
| `retry` | Intelligent retry with automated fixes for common issues |
| `cache-miss` | Analyze cache miss reasons and optimization opportunities |
| `sandbox-debug` | Debug sandbox-related build issues |
| `profile` | Build performance analysis and optimization |
| `watch <package>` | Monitor builds in real-time with AI insights |
| `status [build-id]` | Check status of background builds |
| `stop <build-id>` | Cancel a running background build |
| `background <pkg>` | Start a build in the background |
| `queue <pkg1> <pkg2>` | Build multiple packages sequentially |

### Global Flags

```sh
--flake             Use flake mode for building
--dry-run           Show what would be built without actually building
--verbose           Show verbose build output
--out-link string   Path where the symlink to the output will be stored
-h, --help          Help for build command
```

---

## Basic Usage

### Simple Build with AI Assistance
```sh
# Build current system with AI assistance
nixai build

# Build specific package with AI analysis
nixai build firefox

# Build flake target with AI monitoring
nixai build .#mypackage --flake
```

### Getting Help
```sh
# Show main build help
nixai build --help

# Show subcommand help
nixai build debug --help
nixai build watch --help
```

---

## Advanced Build Analysis

### 🔍 Debug Build Failures
Deep analysis of build failures with AI-powered pattern recognition:

```sh
# Analyze firefox build failure
nixai build debug firefox

# Debug with verbose output
nixai build debug firefox --verbose

# Debug flake build issues
nixai build debug .#myapp --flake
```

**Features:**
- Root cause identification
- Error pattern classification
- Step-by-step fix recommendations
- Alternative solution approaches
- Prevention tips and best practices

### 🔄 Intelligent Retry
Automated retry with smart fixes for common build issues:

```sh
# Retry last failed build with AI fixes
nixai build retry
```

**Automated Fixes Include:**
- Garbage collection cleanup
- Channel updates
- Failed path clearing
- Dependency resolution
- Environment adjustments

### 📊 Performance Profiling
Analyze build performance and identify optimization opportunities:

```sh
# Profile specific package build
nixai build profile --package firefox

# Profile with alternative syntax  
nixai build profile --package vim

# Profile current system build
nixai build profile
```

**Analysis Includes:**
- Build time breakdown
- Resource utilization
- Dependency analysis
- Parallelization opportunities
- System optimization recommendations

---

## Real-Time Monitoring

### 👀 Watch Builds Live
Monitor builds in real-time with AI-powered insights:

```sh
# Watch firefox build with live analysis
nixai build watch firefox

# Watch flake build
nixai build watch .#myapp --flake
```

**Real-Time Features:**
- Live progress updates
- AI error analysis as issues occur
- Performance monitoring
- Intelligent failure recovery suggestions
- Background build management

### 📋 Build Status Management
Check and manage background builds:

```sh
# Show all active builds
nixai build status

# Show specific build status
nixai build status build-12345

# Stop a running build
nixai build stop build-12345
```

---

## Background Build Management

### 🚀 Background Builds
Start builds in background for long-running processes:

```sh
# Start firefox build in background
nixai build background firefox

# Background flake build
nixai build background .#myapp --flake
```

**Benefits:**
- Continue working while building
- AI-powered monitoring
- Status tracking and reporting
- Error analysis and recovery

### 📝 Build Queue Management
Queue multiple packages for sequential building:

```sh
# Queue multiple packages
nixai build queue firefox vim git

# Queue with AI optimization
nixai build queue pkg1 pkg2 pkg3 --flake
```

**AI Optimization Features:**
- Dependency-based build ordering
- Resource usage optimization
- Parallel build coordination
- Progress consolidation

---

## Troubleshooting & Optimization

### 💾 Cache Analysis
Analyze and optimize binary cache performance:

```sh
# Analyze cache performance
nixai build cache-miss
```

**Analysis Includes:**
- Cache hit/miss patterns
- Performance bottlenecks
- Configuration recommendations
- Binary cache optimization
- Network cache strategies

### 🛡️ Sandbox Debugging
Debug sandbox permission and environment issues:

```sh
# Analyze sandbox environment
nixai build sandbox-debug
```

**Troubleshooting:**
- Permission analysis
- Environment restrictions
- Security policy evaluation
- Configuration recommendations
- Resolution procedures

---

## AI-Powered Recovery System

### Intelligent Build Recovery
The build system includes an advanced AI-powered recovery mechanism:

**Features:**
- Pattern recognition for common failures
- Automated fix suggestion and application
- Learning from successful recoveries
- Contextual error analysis
- Progressive retry strategies

**Recovery Process:**
1. Failure detection and classification
2. AI analysis of error patterns
3. Automated fix application
4. Intelligent retry with optimizations
5. Learning integration for future builds

### Build History Tracking
Automatic tracking of build failures and recoveries:

```sh
# Location: ~/.cache/nixai/build-history/
# Format: failure-YYYY-MM-DD-HH-MM-SS.log
```

---

## Real-World Examples

### Basic Build Scenarios
```sh
# Simple package build with AI assistance
nixai build curl

# System rebuild with monitoring
nixai build --verbose

# Flake-based application build
nixai build .#myapp --flake
```

### Advanced Troubleshooting
```sh
# Debug failing Firefox build
nixai build debug firefox

# Analyze why builds aren't using cache
nixai build cache-miss

# Profile performance of complex build
nixai build profile --package nixos-system
```

### Production Workflows
```sh
# Start background build and monitor
nixai build background firefox
nixai build watch firefox

# Queue system rebuild with dependencies
nixai build queue system-config user-packages development-tools

# Monitor all active builds
nixai build status
```

### Emergency Recovery
```sh
# Intelligent retry of failed build
nixai build retry

# Debug sandbox issues preventing builds
nixai build sandbox-debug

# Quick performance check
nixai build profile --package failing-package
```

---

## Best Practices

### For Development
- Use `nixai build watch` for active development
- Profile builds regularly with `nixai build profile`
- Queue related packages with `nixai build queue`

### For Production
- Start critical builds with `nixai build background`
- Monitor build status with `nixai build status`
- Analyze cache performance periodically

### For Troubleshooting
- Always start with `nixai build debug <package>`
- Use `nixai build retry` for known issues
- Check sandbox with `nixai build sandbox-debug`

---

## Integration with nixai Ecosystem

The build system integrates seamlessly with other nixai components:
- **Configuration Management**: Auto-detects NixOS config paths
- **AI Providers**: Supports all configured AI providers (Ollama, OpenAI, Gemini)
- **Logging**: Comprehensive logging with configurable levels
- **Error Recovery**: Learns from `nixai diagnose` and `nixai doctor` results
- **Package Analysis**: Integrates with `nixai package-repo` for dependency insights
