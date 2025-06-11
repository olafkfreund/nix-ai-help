# nixai TUI Modernization Project Plan

## 🎯 Project Overview

**Objective**: Modernize the nixai Terminal User Interface (TUI) to create a cleaner, more accessible, and feature-rich experience without icons, with enhanced typography, improved scrolling, version display, and changelog functionality.

**Target Date**: June 2025
**Status**: Phase 5 Complete (85% Complete) ✅🚧

---

## 📋 Current State Analysis

### Existing TUI Architecture
- **Modern Implementation**: `/internal/tui/` using Bubble Tea framework
- **Legacy Implementation**: `/internal/cli/interactive_tui.go` (1400+ lines)
- **Panel System**: Commands panel (30%), Execution panel (70%), Status bar
- **Styling**: lipgloss-based theming with multiple color schemes
- **Search**: Type-to-search functionality with "/" trigger
- **Focus Management**: Tab-based panel switching

### Current Issues Identified
1. **Icon Dependency**: Extensive use of Unicode icons (🤖, 🔍, 📝, etc.)
2. **Typography**: Left panel text could be larger and more prominent
3. **Scrolling**: Limited scrolling implementation in some panels
4. **Version Info**: No version display in TUI
5. **Feature Discovery**: No changelog or feature announcement system
6. **Accessibility**: Icons may not be accessible to all users

---

## 🎨 Design Goals

### Visual Improvements
- **Icon-Free Interface**: Complete removal of all Unicode icons
- **Enhanced Typography**: Larger, more readable text in command panel
- **Background Styling**: Rich background colors and styling for left panel
- **Modern Layout**: Clean, professional appearance with improved spacing
- **Consistent Theming**: Unified color scheme throughout interface

### Functionality Enhancements
- **Universal Scrolling**: Proper scrolling in all content areas
- **Version Display**: nixai version prominently shown at bottom
- **Changelog Popup**: Scrollable popup showing new features and updates
- **Improved Navigation**: Enhanced keyboard shortcuts and help system
- **Better Accessibility**: Text-based interface without reliance on icons

---

## 🏗️ Implementation Plan

### Phase 1: Icon Removal & Command System Updates
**Estimated Time**: 2-3 days

#### 1.1 Update Command Definitions
**Files to Modify**:
- `/internal/tui/models/command.go`
- `/internal/cli/interactive_tui.go`

**Tasks**:
- Remove `Icon` field from `CommandMetadata` struct
- Update all command definitions to remove icon references
- Modify command rendering to use text-only approach
- Create text-based visual hierarchy (prefixes, indentation)

**Example Changes**:
```go
// Before
{
    Name:        "ask",
    Icon:        "🤖",
    Description: "Ask AI questions about NixOS",
}

// After
{
    Name:        "ask",
    Category:    "AI",
    Description: "Ask AI questions about NixOS",
    Priority:    1, // For visual ordering
}
```

#### 1.2 Update Status Indicators
**Files to Modify**:
- `/internal/tui/panels/status.go`
- `/internal/tui/panels/execution.go`

**Tasks**:
- Replace icon-based status indicators with text
- Use color coding instead of icons for status
- Create text-based progress indicators

### Phase 2: Enhanced Typography & Styling
**Estimated Time**: 2-3 days

#### 2.1 Enhance Commands Panel Typography
**Files to Modify**:
- `/internal/tui/panels/commands.go`
- `/internal/tui/styles/theme.go`

**Tasks**:
- Increase font weight/boldness for command names
- Add background styling for selected items
- Implement larger text rendering for command panel
- Create visual hierarchy with spacing and indentation

**Styling Changes**:
```go
// Enhanced command panel styles
CommandsPanel: PanelStyles{
    Base: lipgloss.NewStyle().
        Background(background).
        Foreground(text).
        Padding(1),
    Header: lipgloss.NewStyle().
        Foreground(primary).
        Bold(true).
        Background(secondary).
        Padding(0, 2).
        MarginBottom(1),
    Selected: lipgloss.NewStyle().
        Background(primary).
        Foreground(background).
        Bold(true).
        Padding(0, 2).
        MarginLeft(1).
        MarginRight(1),
    Content: lipgloss.NewStyle().
        Foreground(text).
        Padding(0, 2).
        MarginLeft(2),
}
```

#### 2.2 Background Styling System
**New Component**: Enhanced background patterns and colors

**Tasks**:
- Create rich background styling for panels
- Implement gradient-like effects using character patterns
- Add subtle borders and separators
- Design focus indicators without icons

### Phase 3: Scrolling Implementation
**Estimated Time**: 2-3 days

#### 3.1 Universal Scrolling System
**Files to Modify**:
- `/internal/tui/panels/commands.go`
- `/internal/tui/panels/execution.go`
- `/internal/tui/app.go`

**Tasks**:
- Implement viewport scrolling in commands panel
- Enhance execution panel scrolling
- Add scroll indicators (text-based)
- Implement smooth scrolling behavior
- Add page up/down functionality

**Scrolling Features**:
- **Commands Panel**: Smooth vertical scrolling through command list
- **Execution Panel**: Auto-scroll to bottom, manual scroll capability
- **Help Text**: Scrollable help and documentation areas
- **Search Results**: Scrollable filtered command lists

#### 3.2 Scroll Indicators
**Implementation**: Text-based scroll position indicators

```go
// Example scroll indicator
scrollInfo := fmt.Sprintf("(%d-%d of %d)", 
    p.scrollOffset+1, 
    min(p.scrollOffset+visibleHeight, totalItems), 
    totalItems)
```

### Phase 4: Version Display System
**Estimated Time**: 1-2 days

#### 4.1 Version Information Component
**New Files**:
- `/internal/tui/components/version.go`

**Files to Modify**:
- `/internal/tui/panels/status.go`
- `/pkg/version/version.go`

**Tasks**:
- Create version display component
- Integrate with existing version system
- Add build information (commit, date)
- Display in status bar area

**Version Display Design**:
```
┌─ Commands ─────────────────┬─ Execution ─────────────────────┐
│                            │                                 │
│                            │                                 │
└────────────────────────────┴─────────────────────────────────┘
  Status Info                    nixai v1.2.3 (abc1234 - Jun 9)
```

#### 4.2 Build Information Integration
**Tasks**:
- Extract version from build flags
- Display git commit hash (short)
- Show build date
- Add development vs release indicators

### Phase 5: Changelog Popup System
**Estimated Time**: 3-4 days

#### 5.1 Changelog Data Management
**New Files**:
- `/internal/tui/components/changelog.go`
- `/configs/changelog.yaml`

**Tasks**:
- Create changelog data structure
- Implement version-based changelog loading
- Design scrollable popup interface
- Add keyboard navigation for popup

**Changelog Structure**:
```yaml
changelog:
  - version: "1.2.3"
    date: "2025-06-09"
    highlights:
      - "Enhanced TUI with modern design"
      - "Removed icon dependencies"
      - "Improved scrolling throughout"
    features:
      - title: "New Search System"
        description: "Faster, more accurate package search"
      - title: "Better Error Handling"
        description: "Clearer error messages and recovery"
    fixes:
      - "Fixed scrolling in commands panel"
      - "Resolved theme switching issues"
```

#### 5.2 Popup Interface Design
**Features**:
- **Trigger**: F1 or Ctrl+? to open changelog
- **Navigation**: Arrow keys, Page Up/Down, Home/End
- **Scrollable**: Full vertical scrolling capability
- **Searchable**: Filter changelog entries
- **Closable**: Escape or Enter to close

**Popup Layout**:
```
┌─ nixai Changelog ─────────────────────────────────────────┐
│                                                           │
│  Version 1.2.3 (June 9, 2025)                           │
│  ✨ New Features:                                        │
│    • Enhanced TUI with modern design                     │
│    • Removed icon dependencies for accessibility         │
│    • Improved scrolling throughout interface             │
│                                                           │
│  🔧 Improvements:                                        │
│    • Faster package search performance                   │
│    • Better error messages and recovery                  │
│                                                           │
│  🐛 Bug Fixes:                                          │
│    • Fixed scrolling in commands panel                   │
│    • Resolved theme switching issues                     │
│                                                           │
│  ─────────────────────────────────────────────────────── │
│                                                           │
│  Version 1.2.2 (June 1, 2025)                           │
│  ...                                                      │
│                                                           │
│  [Esc] Close  [↑↓] Scroll  [PgUp/PgDn] Page  [/] Search │
└───────────────────────────────────────────────────────────┘
```

### Phase 6: Modern UI Polish
**Estimated Time**: 2-3 days

#### 6.1 Layout Optimization
**Tasks**:
- Adjust panel proportions for better balance
- Implement responsive layout for different terminal sizes
- Add padding and spacing improvements
- Create visual separators without using special characters

#### 6.2 Keyboard Shortcuts Enhancement
**New Shortcuts**:
- `F1` or `Ctrl+?`: Open changelog popup
- `F2`: Toggle theme
- `F3`: Show version info
- `F5`: Refresh/reload
- `Ctrl+D`: Open documentation
- `Ctrl+S`: Quick save/bookmark

#### 6.3 Help System Updates
**Tasks**:
- Update all help text to reflect new shortcuts
- Remove icon references from help
- Add context-sensitive help
- Create quick reference popup

### Phase 7: Testing & Validation
**Estimated Time**: 2-3 days

#### 7.1 Functionality Testing
**Test Areas**:
- All panel interactions without icons
- Scrolling in all areas
- Version display accuracy
- Changelog popup functionality
- Theme switching
- Keyboard navigation
- Search functionality
- Command execution

#### 7.2 Accessibility Testing
**Requirements**:
- Text-only interface compatibility
- Screen reader friendly
- High contrast theme support
- Keyboard-only navigation
- No dependency on Unicode icons

#### 7.3 Performance Testing
**Metrics**:
- TUI startup time
- Scrolling responsiveness
- Memory usage
- CPU usage during interaction

---

## 📂 File Modification Plan

### Core Files to Modify

#### 1. Command System
```
internal/tui/models/command.go        - Remove Icon field, add visual hierarchy
internal/cli/interactive_tui.go       - Update legacy TUI icon removal
```

#### 2. Panel System
```
internal/tui/panels/commands.go       - Enhanced typography, scrolling
internal/tui/panels/execution.go      - Improved scrolling, text indicators
internal/tui/panels/status.go         - Version display, text-based status
```

#### 3. Styling System
```
internal/tui/styles/theme.go          - Enhanced typography styles
internal/tui/app.go                   - Layout adjustments
```

#### 4. New Components
```
internal/tui/components/version.go    - Version display component
internal/tui/components/changelog.go  - Changelog popup system
internal/tui/components/popup.go      - Generic popup framework
```

#### 5. Configuration
```
configs/changelog.yaml               - Changelog data
configs/default.yaml                - TUI configuration options
```

### Supporting Files
```
pkg/version/version.go               - Version information utilities
internal/tui/models/app_state.go     - State management updates
```

---

## 🎯 Success Criteria

### Visual Requirements ✅
- [ ] Complete removal of all Unicode icons
- [ ] Enhanced typography in commands panel (larger, bolder text)
- [ ] Rich background styling throughout interface
- [ ] Modern, clean appearance
- [ ] Consistent theming

### Functionality Requirements ✅
- [ ] Universal scrolling in all panels
- [ ] Version display prominently shown
- [ ] Changelog popup accessible via keyboard shortcut
- [ ] All existing functionality preserved
- [ ] Improved keyboard navigation

### Accessibility Requirements ✅
- [ ] Text-only interface (no icon dependencies)
- [ ] Screen reader compatibility
- [ ] High contrast theme support
- [ ] Keyboard-only navigation capability

### Performance Requirements ✅
- [ ] TUI startup time < 1 second
- [ ] Smooth scrolling performance
- [ ] Memory usage optimized
- [ ] Responsive user interactions

---

## 🔧 Implementation Commands

### Development Setup
```bash
# Enter development environment
cd /home/olafkfreund/Source/NIX/nix-ai-help
nix develop

# Start development workflow
just build-watch    # Continuous build during development
just test-watch     # Continuous testing
```

### Testing Commands
```bash
# Test TUI specifically
just test-tui

# Test with different themes
./nixai --tui --theme dark
./nixai --tui --theme light
./nixai --tui --theme nixos

# Test accessibility
./nixai --tui --no-color
./nixai --tui --high-contrast
```

### Validation Commands
```bash
# Version display test
./nixai --version
./nixai --tui  # Check version in status bar

# Changelog test
./nixai --tui  # Press F1 for changelog

# Scrolling test
./nixai --tui  # Navigate through commands, test all scrolling
```

---

## 📈 Timeline & Milestones

### Week 1: Core Implementation
- **Days 1-2**: Icon removal and command system updates
- **Days 3-4**: Enhanced typography and styling
- **Days 5-7**: Scrolling implementation

### Week 2: Advanced Features
- **Days 1-2**: Version display system
- **Days 3-5**: Changelog popup system
- **Days 6-7**: UI polish and optimization

### Week 3: Testing & Refinement
- **Days 1-3**: Comprehensive testing and bug fixes
- **Days 4-5**: Accessibility testing and improvements
- **Days 6-7**: Performance optimization and final polish

---

## 🚀 Future Enhancements

### Phase 2 Features (Post-Launch)
- **Custom Themes**: User-configurable color schemes
- **Layout Options**: Different panel arrangements
- **Plugin System**: Extensible TUI components
- **Saved Sessions**: Remember user preferences
- **Command History**: Enhanced command history with search
- **Bookmarks**: Save favorite commands and configurations

### Advanced Features
- **Split Panels**: Multiple execution panels
- **Tabs**: Tabbed interface for multiple sessions
- **Minimap**: Command overview panel
- **Quick Actions**: Configurable quick action bar
- **Context Menus**: Right-click equivalent actions

---

## 📝 Notes & Considerations

### Technical Considerations
- **Bubble Tea Framework**: Leverage existing framework capabilities
- **Backward Compatibility**: Ensure legacy TUI still functions
- **Configuration**: Make new features configurable
- **Performance**: Optimize for large command lists
- **Memory**: Efficient changelog and version caching

### User Experience
- **Migration**: Smooth transition for existing users
- **Learning Curve**: Maintain familiar navigation patterns
- **Documentation**: Update all documentation to reflect changes
- **Feedback**: Implement user feedback collection

### Testing Strategy
- **Unit Tests**: Individual component testing
- **Integration Tests**: Full TUI workflow testing
- **User Testing**: Real-world usage scenarios
- **Accessibility Testing**: Screen reader and keyboard-only testing

---

## 🎉 Expected Outcomes

Upon completion of this modernization project, users will experience:

1. **Cleaner Interface**: Modern, professional appearance without clutter
2. **Better Accessibility**: Text-only interface accessible to all users
3. **Enhanced Usability**: Larger text, better contrast, improved navigation
4. **Feature Discovery**: Easy access to changelog and new features
5. **Professional Feel**: Enterprise-ready terminal interface
6. **Improved Performance**: Smooth scrolling and responsive interactions

This modernization will position nixai as a leading example of accessible, modern terminal user interface design while maintaining all existing functionality and improving the overall user experience.

---

*Last Updated: June 9, 2025*
*Project Phase: Planning*
*Next Review: Upon Phase 1 Completion*
