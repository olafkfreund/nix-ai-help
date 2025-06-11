# Context-Aware NixOS Suggestions Implementation Plan

## 🎯 Project Overview

Enhance nixai to automatically detect and understand the user's NixOS configuration context (flakes vs channels, Home Manager setup, enabled services, etc.) to provide more accurate, personalized suggestions.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Input    │    │  Context        │    │  AI Provider    │
│   (Questions)   │───▶│  Detection      │───▶│  (Enhanced)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Configuration  │
                       │  Analysis       │
                       └─────────────────┘
```

## 📋 Implementation Phases

### Phase 1: Configuration Detection Infrastructure

- **Duration**: 3-4 days
- **Files**: `internal/nixos/context_detector.go`, `internal/config/config.go`
- **Goal**: Implement comprehensive NixOS configuration detection

#### Tasks

1. ✅ Extend `UserConfig` struct with `NixOSContext`
2. ✅ Create `DetectNixOSContext()` function
3. ✅ Implement system type detection (NixOS/nix-darwin/home-manager-only)
4. ✅ Implement flakes vs channels detection
5. ✅ Implement Home Manager detection (standalone/module/none)
6. ✅ Implement configuration file parsing
7. ✅ Create caching mechanism for detected context

### Phase 2: Context-Aware AI Integration

- **Duration**: 2-3 days  
- **Files**: `internal/ai/context/nixos_context.go`, `internal/cli/direct_commands.go`
- **Goal**: Integrate context detection with AI providers

#### Tasks

1. ✅ Create `BuildContextualPrompt()` function
2. ✅ Update ask command to use contextual prompts
3. ✅ Implement context caching and refresh logic
4. ✅ Add context summary display to users
5. ✅ Test with different configuration scenarios

### Phase 3: Context Management Commands

- **Duration**: 2 days
- **Files**: `internal/cli/context_commands.go`
- **Goal**: Provide user control over context detection

#### Tasks

1. ⏳ Create `nixai context detect` command
2. ⏳ Create `nixai context show` command  
3. ⏳ Create `nixai context reset` command
4. ⏳ Add context validation and health checks
5. ⏳ Implement automatic context refresh triggers

### Phase 4: Enhanced Suggestion Logic

- **Duration**: 2-3 days
- **Files**: Various AI provider files, `internal/ai/agents/`
- **Goal**: Implement context-specific suggestion logic

#### Tasks

1. ⏳ Update all AI agents to use contextual prompts
2. ⏳ Create configuration-specific templates
3. ⏳ Implement smart fallback suggestions
4. ⏳ Add configuration validation warnings
5. ⏳ Test across all command categories

### Phase 5: Testing & Documentation

- **Duration**: 2 days
- **Files**: `tests/`, `docs/`, `README.md`
- **Goal**: Comprehensive testing and documentation

#### Tasks

1. ⏳ Create unit tests for context detection
2. ⏳ Create integration tests for various setups
3. ⏳ Update documentation and examples
4. ⏳ Create troubleshooting guide
5. ⏳ Performance optimization

## 🏷️ Current Status: Phase 1 - In Progress

## 📊 Technical Specifications

### Configuration Context Structure

```go
type NixOSContext struct {
    // System Detection
    UsesFlakes           bool     
    UsesChannels         bool     
    NixOSConfigPath      string   
    SystemType           string   // "nixos", "nix-darwin", "home-manager-only"
    
    // Home Manager
    HasHomeManager       bool     
    HomeManagerType      string   // "standalone", "module", "none"
    HomeManagerConfigPath string  
    
    // Version Information  
    NixOSVersion         string   
    NixVersion           string   
    
    // Configuration Analysis
    ConfigurationFiles   []string 
    EnabledServices      []string 
    InstalledPackages    []string 
    
    // File Paths
    FlakeFile           string   
    ConfigurationNix    string   
    HardwareConfigNix   string   
}
```

### Detection Priority Order

1. **User-specified paths** (from user config)
2. **Common NixOS locations** (`/etc/nixos`, `~/.config/nixos`)
3. **Home Manager locations** (`~/.config/home-manager`)
4. **Git repository detection** (for flake-based setups)
5. **Fallback to system defaults**

### Context-Aware Prompt Examples

#### Flakes + Home Manager (Module)

```
=== USER'S NIXOS CONTEXT ===
System Type: nixos
✅ USES FLAKES - Always suggest flake-based solutions
❌ NEVER suggest nix-channel commands
✅ HAS HOME MANAGER AS NIXOS MODULE
Use home-manager.users.<username> syntax
Currently enabled services: nginx, postgresql
=== END CONTEXT ===
```

#### Traditional Channels Only

```
=== USER'S NIXOS CONTEXT ===
System Type: nixos  
Uses legacy channels - suggest channel-compatible solutions
❌ NO HOME MANAGER - Only suggest system-level configuration
Currently enabled services: sshd, firewall
=== END CONTEXT ===
```

## 🎯 Success Criteria

### User Experience Improvements

- ✅ **Accurate Suggestions**: Suggestions match user's actual setup
- ⏳ **No Manual Configuration**: Context detection works automatically  
- ⏳ **Clear Feedback**: Users understand what context is detected
- ⏳ **Performance**: Context detection completes under 2 seconds
- ⏳ **Reliability**: 95%+ accuracy in configuration detection

### Technical Achievements

- ✅ **Comprehensive Detection**: Covers all major NixOS setup patterns
- ⏳ **Efficient Caching**: Context cached and refreshed appropriately
- ⏳ **Error Handling**: Graceful degradation when detection fails
- ⏳ **Extensibility**: Easy to add new context detection methods
- ⏳ **Testing Coverage**: >90% test coverage for context detection

## 🧪 Testing Strategy

### Unit Tests

- ✅ Configuration file parsing accuracy
- ⏳ System type detection logic
- ⏳ Home Manager detection scenarios
- ⏳ Flakes vs channels detection
- ⏳ Context caching mechanisms

### Integration Tests  

- ⏳ End-to-end suggestion accuracy across setups
- ⏳ Performance benchmarks for context detection
- ⏳ Error handling with malformed configurations
- ⏳ Cross-platform compatibility (NixOS, nix-darwin)

### Manual Testing Scenarios

- ⏳ Fresh NixOS installation with flakes
- ⏳ Traditional channel-based setup
- ⏳ Home Manager standalone installation
- ⏳ Home Manager as NixOS module
- ⏳ Mixed environments and edge cases

## 📈 Performance Considerations

### Optimization Strategies

- ⏳ **Lazy Loading**: Detect context only when needed
- ⏳ **Parallel Detection**: Run multiple detection methods concurrently
- ⏳ **Smart Caching**: Cache based on file modification times
- ⏳ **Incremental Updates**: Update only changed context elements

### Performance Targets

- ⏳ Initial context detection: < 2 seconds
- ⏳ Cached context retrieval: < 100ms
- ⏳ Context refresh: < 1 second
- ⏳ Memory usage: < 10MB additional overhead

## 🚀 Future Enhancements

### Phase 6+ (Future Iterations)

- ⏳ **Machine Learning**: Learn from user corrections and preferences
- ⏳ **Remote Context**: Detect context from remote Git repositories
- ⏳ **Team Setups**: Support for shared team configurations
- ⏳ **Configuration Migration**: Automated migration suggestions
- ⏳ **Advanced Analytics**: Configuration health scoring and recommendations

## 📚 Resources

### Documentation Links

- [NixOS Manual](https://nixos.org/manual/nixos/stable/)
- [Home Manager Manual](https://nix-community.github.io/home-manager/)
- [Nix Flakes](https://nixos.wiki/wiki/Flakes)
- [NixOS Configuration](https://nixos.wiki/wiki/Configuration)

### Code References

- `internal/config/config.go` - Configuration management
- `internal/ai/` - AI provider integration
- `pkg/utils/` - Utility functions
- `internal/cli/` - Command-line interface

---

**Last Updated**: December 2024  
**Status**: Phase 1 - In Progress  
**Next Milestone**: Complete context detection infrastructure
