# NixAI Interactive Mode TUI Rewrite Project Plan

## Overview

This document outlines the comprehensive project plan for rewriting the nixai interactive mode into a modern Terminal User Interface (TUI) that provides a more intuitive and efficient command-line experience.

## Current State Analysis

### Existing Interactive Mode
- **Location**: `internal/cli/interactive.go`
- **Interface**: Simple command prompt (`nixai> `)
- **Completion**: Basic readline completion via `interactive_completion.go`
- **Commands**: 29+ available commands (ask, build, community, diagnose, doctor, etc.)
- **Execution**: Direct command execution through `RunDirectCommand`
- **Dependencies**: Uses `github.com/chzyer/readline` for input handling

### Current Limitations
- No visual command discovery
- Limited command context
- No persistent command history view
- No side-by-side execution monitoring
- Basic text-only interface
- Users must remember all command names

## Target TUI Design

### Visual Layout
```
┌─ Commands (30%)  ──────────┬─ Execution Area (70%)  ────────────────────┐
│ 🔍 [Search: type to filter │ Command Output:                            │
│                            │ ┌────────────────────────────────────────┐ │
│ 🤖 ask                     │ │ $ nixai doctor --verbose               │ │
│ 🛠️ build                   │ │ ✅ NixOS Health Check Results          │ │
│ 🌐 community               │ │ System: OK                             │ │
│ ⚙️ config                  │ │ Configuration: OK                      │ │
│ 🧑‍💻 configure               │ │ Dependencies: 2 warnings               │ │
│ 🩺 diagnose                │ │ ⚠️  Missing: git, curl                 │ │
│ 🩻 doctor                  │ │                                        │ │
│ 🖥️ explain-option          │ │ Command completed in 1.2s              │ │
│ 🧊 flake                   │ └────────────────────────────────────────┘ │
│ 🧹 gc                      │                                            │
│ 💻 hardware                │ Command Input:                             │
│ 📚 learn                   │ ┌────────────────────────────────────────┐ │
│ 📝 logs                    │ │ explain-option services.openssh        │ │
│ 🖧 machines                │ └────────────────────────────────────────┘ │
│ 🛰️ mcp-server              │ [Execute] [Clear] [History ▼] [Help]       │
│ 🔀 migrate                 │                                            │
│ 📦 package-repo            │ ⏳ Command History:                        │
│ 🔍 search                  │ • doctor --verbose                         │
│ 🔖 snippets                │ • ask "how to enable SSH"                  │
│ 💾 store                   │ • build --check                            │
│ 📄 templates               │                                            │
│ ❌ exit                    │                                            │
└────────────────────────────┴────────────────────────────────────────────┘
│ Status:● MCP Running ● AI:Ollama ● NixOS:/etc/nixos │ F1:Help │ Ctrl+C:Exit │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Key Features
1. **Left Panel**: Searchable command list with icons and descriptions
2. **Right Panel**: Command execution area with scrollable output
3. **Status Bar**: System status indicators and key bindings
4. **Enhanced Input**: Auto-completion, history, syntax highlighting
5. **Real-time Filtering**: Type-to-search command functionality
6. **Multi-session Support**: Tabbed interface for multiple command sessions

## Technical Architecture

### New Package Structure
```
internal/tui/
├── app.go                 # Main TUI application controller
├── panels/
│   ├── commands.go        # Left sidebar commands panel
│   ├── execution.go       # Right execution panel
│   └── status.go          # Bottom status bar
├── components/
│   ├── command_input.go   # Enhanced input component
│   ├── output_view.go     # Scrollable output display
│   ├── search.go          # Command search/filter
│   └── help.go            # Help overlay
├── models/
│   ├── app_state.go       # Application state management
│   ├── command.go         # Command metadata structures
│   └── session.go         # Command session handling
├── styles/
│   ├── theme.go           # Color themes and styling
│   ├── layout.go          # Layout definitions
│   └── components.go      # Component-specific styles
├── services/
│   ├── executor.go        # Command execution service
│   ├── completion.go      # Enhanced completion service
│   └── history.go         # Command history management
└── utils/
    ├── keys.go            # Key binding definitions
    ├── formatter.go       # Output formatting utilities
    └── discovery.go       # Command discovery from cobra
```

## Dependencies

### New Dependencies to Add
```go
// Core TUI framework
github.com/charmbracelet/bubbletea v0.24.2

// UI components and styling
github.com/charmbracelet/lipgloss v0.9.1
github.com/charmbracelet/bubbles v0.16.1

// Additional utilities
github.com/muesli/termenv v0.15.2
github.com/lucasb-eyer/go-colorful v1.2.0
```

### Existing Dependencies to Leverage
- `github.com/charmbracelet/glamour` (already in use for markdown)
- `github.com/charmbracelet/lipgloss` (already in use for styling)
- `github.com/chzyer/readline` (keep for backward compatibility)

## Implementation Phases

### Phase 1: Foundation Setup (Week 1)

#### Objectives
- Set up TUI framework and basic structure
- Create main application skeleton
- Implement basic two-panel layout

#### Tasks
1. **Dependencies Setup**
   ```bash
   go get github.com/charmbracelet/bubbletea@v0.24.2
   go get github.com/charmbracelet/bubbles@v0.16.1
   ```

2. **Core Structure Creation**
   - Create `internal/tui/` package
   - Implement `app.go` with basic bubbletea model
   - Set up panel structure with lipgloss styling

3. **Basic Layout Implementation**
   - Two-column layout with adjustable split
   - Command list in left panel
   - Empty execution area in right panel

4. **Integration Point**
   - Modify `internal/cli/interactive.go` to add `--tui` flag
   - Create `InteractiveModeTUI()` function

#### Deliverables
- Basic TUI application that starts and displays layout
- Command list populated from existing commands
- Navigation between panels with Tab/Shift+Tab

### Phase 2: Command Discovery & Integration (Week 2)

#### Objectives
- Integrate with existing command system
- Implement command execution in TUI
- Add basic output display

#### Tasks
1. **Command Discovery Service**
   ```go
   // internal/tui/services/discovery.go
   func DiscoverCommands() []CommandMetadata
   func GetCommandHelp(name string) string
   func GetSubcommands(name string) []string
   ```

2. **Command Execution Integration**
   - Enhance `RunDirectCommand` for TUI compatibility
   - Implement streaming output capture
   - Add progress indicators for long-running commands

3. **Basic Output Display**
   - Scrollable output view
   - Command execution status
   - Error highlighting

4. **Core Commands Testing**
   - Test with `ask`, `doctor`, `diagnose` commands
   - Ensure AI provider integration works
   - Verify MCP server communication

#### Deliverables
- Working command execution within TUI
- Scrollable output display
- Integration with core nixai functionality

### Phase 3: Enhanced User Experience (Week 3)

#### Objectives
- Implement advanced completion and search
- Add command history management
- Enhance input experience

#### Tasks
1. **Advanced Command Completion**
   ```go
   // internal/tui/services/completion.go
   func GetCompletions(partial string, context CommandContext) []Completion
   func GetParameterSuggestions(cmd string, param string) []string
   ```

2. **Real-time Search/Filter**
   - Type-to-search in command panel
   - Fuzzy matching for command discovery
   - Category-based filtering

3. **Command History System**
   - Persistent history across sessions
   - History navigation with arrow keys
   - History search and filtering

4. **Enhanced Input Component**
   - Syntax highlighting for commands
   - Multi-line input support
   - Smart parameter completion

#### Deliverables
- Real-time command search and filtering
- Advanced tab completion system
- Persistent command history
- Enhanced input experience

### Phase 4: Advanced Features (Week 4)

#### Objectives
- Add multi-session support
- Implement themes and customization
- Add help system and documentation

#### Tasks
1. **Multi-Session Support**
   - Tabbed interface for multiple command sessions
   - Session state management
   - Background command execution

2. **Theming System**
   ```go
   // internal/tui/styles/theme.go
   type Theme struct {
       Primary     lipgloss.Color
       Secondary   lipgloss.Color
       Accent      lipgloss.Color
       Background  lipgloss.Color
       Text        lipgloss.Color
   }
   ```

3. **Help System**
   - Contextual help overlay (F1)
   - Command-specific help
   - Key binding reference

4. **Status Bar Enhancement**
   - MCP server status indicator
   - AI provider status
   - Current NixOS configuration path
   - Resource usage indicators

#### Deliverables
- Tabbed multi-session interface
- Multiple color themes
- Comprehensive help system
- Rich status indicators

### Phase 5: Polish & Testing (Week 5)

#### Objectives
- Comprehensive testing and bug fixes
- Performance optimization
- Documentation and user guides

#### Tasks
1. **Testing Suite**
   - Unit tests for TUI components
   - Integration tests with existing commands
   - Performance benchmarks

2. **User Experience Polish**
   - Responsive design for different terminal sizes
   - Smooth animations and transitions
   - Error handling and recovery

3. **Documentation**
   - Update `docs/interactive.md`
   - Create TUI user guide
   - Update README with new features

4. **Migration Strategy**
   - Backward compatibility testing
   - Migration guide for existing users
   - Feature comparison documentation

#### Deliverables
- Comprehensive test suite
- Performance-optimized TUI
- Complete documentation
- Migration guide

## Key Bindings Design

### Global Navigation
```
Tab / Shift+Tab    - Switch between panels
Ctrl+C             - Exit application
Ctrl+H / F1        - Show help overlay
Ctrl+T             - Toggle theme
Ctrl+N             - New session tab
Ctrl+W             - Close current session
Ctrl+1-9           - Switch to session tab N
```

### Commands Panel
```
↑ / ↓              - Navigate command list
Enter              - Select/execute command
/                  - Start search/filter mode
Esc                - Clear search, return to list
Home / End         - Jump to first/last command
Page Up/Down       - Fast scroll through commands
```

### Execution Panel
```
Enter              - Execute current command
↑ / ↓              - Navigate command history
Tab                - Auto-complete current input
Ctrl+L             - Clear output area
Ctrl+K             - Clear input field
Ctrl+R             - Search command history
Page Up/Down       - Scroll output
Ctrl+F             - Search in output
```

### Advanced Features
```
Alt+←/→            - Resize panels
Ctrl+S             - Save session/output
Ctrl+O             - Open saved session
F2                 - Rename current session
F5                 - Refresh command list
```

## Integration Points

### Existing Command System
```go
// Enhance internal/cli/direct_commands.go
func RunDirectCommandTUI(cmd string, args []string) (*CommandResult, error) {
    // Return structured result for TUI consumption
    return &CommandResult{
        Output:    output,
        Error:     err,
        ExitCode:  code,
        Duration:  elapsed,
        Streaming: isStreaming,
    }, nil
}
```

### AI Provider Integration
```go
// Enhance internal/ai/ for TUI streaming
func (p *Provider) QueryStream(question string) (<-chan string, error) {
    // Return streaming channel for real-time AI responses
}
```

### MCP Server Integration
```go
// Add TUI-specific MCP methods
func (c *Client) GetDocumentationStream(query string) (<-chan string, error) {
    // Stream documentation results for TUI display
}
```

## Configuration Integration

### TUI-Specific Configuration
```yaml
# configs/default.yaml additions
tui:
  enabled: true
  theme: "default"  # default, dark, light, nixos
  panel_ratio: 0.3  # Left panel width ratio
  max_history: 1000
  auto_complete: true
  search_fuzzy: true
  animations: true
  status_bar: true
  multi_session: true
  session_persistence: true
```

### Theme Configuration
```yaml
themes:
  default:
    primary: "#5c7cfa"
    secondary: "#495057"
    accent: "#51cf66"
    background: "#1a1b26"
    text: "#a9b1d6"
  nixos:
    primary: "#7ebae4"
    secondary: "#414868"
    accent: "#9ece6a"
    background: "#24283b"
    text: "#c0caf5"
```

## Backward Compatibility Strategy

### Dual Mode Support
1. **Default Behavior**: Keep existing interactive mode as default
2. **Opt-in TUI**: Use `nixai interactive --tui` to enable new interface
3. **Environment Variable**: `NIXAI_TUI=1` for default TUI mode
4. **Configuration Setting**: `tui.enabled: true` in config file

### Migration Path
1. **Phase 1**: Introduce TUI as optional feature
2. **Phase 2**: Make TUI default with fallback option
3. **Phase 3**: Deprecate old interactive mode (with notice period)
4. **Phase 4**: Remove old mode (major version bump)

## Testing Strategy

### Unit Testing
```go
// internal/tui/app_test.go
func TestAppInitialization(t *testing.T) {}
func TestPanelNavigation(t *testing.T) {}
func TestCommandExecution(t *testing.T) {}

// internal/tui/services/completion_test.go
func TestCommandCompletion(t *testing.T) {}
func TestParameterSuggestions(t *testing.T) {}

// internal/tui/components/search_test.go
func TestCommandSearch(t *testing.T) {}
func TestFuzzyMatching(t *testing.T) {}
```

### Integration Testing
```bash
# Test TUI with all major commands
go test -tags=integration ./internal/tui/...

# Test TUI performance with large outputs
go test -bench=. ./internal/tui/...

# Test TUI in different terminal environments
TERM=xterm-256color go test ./internal/tui/...
```

### Manual Testing Checklist
- [ ] All commands executable from TUI
- [ ] Command completion works correctly
- [ ] Search/filter functionality
- [ ] Multi-session support
- [ ] Theme switching
- [ ] Help system accessibility
- [ ] Performance with large outputs
- [ ] Responsive design in different terminal sizes
- [ ] Error handling and recovery

## Performance Considerations

### Optimization Targets
1. **Startup Time**: TUI should initialize in < 100ms
2. **Command Execution**: No noticeable delay over direct execution
3. **Output Rendering**: Smooth scrolling for large outputs
4. **Memory Usage**: Efficient handling of command history and output

### Implementation Strategies
1. **Lazy Loading**: Load command metadata on demand
2. **Virtual Scrolling**: For large output displays
3. **Background Processing**: Non-blocking command execution
4. **Output Buffering**: Efficient streaming for real-time output

## Documentation Updates Required

### User Documentation
1. **`docs/interactive.md`**: Complete rewrite with TUI features
2. **`docs/MANUAL.md`**: Add TUI section and key bindings
3. **`README.md`**: Update screenshots and feature list

### Developer Documentation
1. **Architecture documentation**: TUI package structure
2. **Contributing guide**: TUI development guidelines
3. **Testing guide**: TUI-specific testing procedures

## Success Metrics

### User Experience Metrics
- Reduced time to discover and execute commands
- Increased user engagement with advanced features
- Positive user feedback on interface improvements

### Technical Metrics
- Command execution performance parity
- Memory usage within acceptable limits
- Comprehensive test coverage (>90%)
- Cross-platform compatibility

## Risk Mitigation

### Technical Risks
1. **Performance Issues**: Mitigate with profiling and optimization
2. **Terminal Compatibility**: Test across major terminal emulators
3. **Dependency Issues**: Pin dependency versions, maintain fallbacks

### User Experience Risks
1. **Learning Curve**: Provide comprehensive help and tutorials
2. **Feature Regression**: Maintain backward compatibility
3. **Accessibility**: Ensure keyboard-only navigation

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|-----------------|
| Phase 1 | Week 1 | Basic TUI structure and layout |
| Phase 2 | Week 2 | Command integration and execution |
| Phase 3 | Week 3 | Advanced UX features |
| Phase 4 | Week 4 | Multi-session and theming |
| Phase 5 | Week 5 | Testing and documentation |

**Total Duration**: 5 weeks

## Future Enhancements

### Post-Launch Features
1. **Visual Diff Viewer**: For configuration changes
2. **Interactive Tutorials**: Built-in learning system
3. **Plugin System**: Third-party command extensions
4. **Remote Sessions**: SSH-based remote nixai execution
5. **AI Chat Interface**: Integrated conversational AI
6. **File Browser**: Navigate and edit NixOS configurations
7. **Log Viewer**: Real-time system log monitoring

### Advanced Integrations
1. **Home Manager TUI**: Dedicated Home Manager interface
2. **Flake Manager**: Visual flake.nix editor
3. **Package Browser**: Interactive package search and install
4. **System Monitor**: Real-time system status and metrics

## Conclusion

This comprehensive rewrite will transform nixai from a simple command-line tool into a modern, efficient TUI application that significantly improves user productivity and command discovery. The phased approach ensures stable development while maintaining backward compatibility and allowing for user feedback integration throughout the process.

The new TUI interface will position nixai as a cutting-edge tool in the NixOS ecosystem, providing users with an intuitive and powerful interface for managing their NixOS systems.
