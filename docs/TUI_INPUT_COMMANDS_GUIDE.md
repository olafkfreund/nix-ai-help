# nixai TUI Input Commands Guide

## Overview

The nixai TUI supports several commands that require input parameters. These commands are marked with `[INPUT]` in the command list and provide interactive ways to ask questions, search packages, analyze repositories, and explain NixOS options.

## Commands That Require Input

### 1. `ask [INPUT]` - AI Question Assistant
**Purpose**: Ask any NixOS-related question to the AI assistant

**Examples**:
- `ask "how do I enable SSH?"`
- `ask "what is the difference between channels and flakes?"`
- `ask "how to configure nvidia drivers?"`
- `ask "best practices for NixOS configuration?"`

**Usage in TUI**:
1. Select `ask [INPUT]` from the command list (or press `a` to search)
2. Tab to execution panel or press Enter to populate command
3. Add your question in quotes after `ask`
4. Press Enter to execute

### 2. `search [INPUT]` - Package and Service Search
**Purpose**: Search for NixOS packages, services, and options

**Examples**:
- `search firefox`
- `search "text editor"`
- `search nginx`
- `search "development tools"`

**Usage in TUI**:
1. Select `search [INPUT]` from the command list
2. Tab to execution panel 
3. Add your search term after `search`
4. Press Enter to execute

**What you get**:
- Complete package list with descriptions and versions
- AI-powered configuration suggestions
- Best practices for installation and setup

### 3. `explain-option [INPUT]` - NixOS Option Explainer
**Purpose**: Get detailed explanations of NixOS configuration options

**Examples**:
- `explain-option services.openssh.enable`
- `explain-option boot.loader.systemd-boot.enable`
- `explain-option networking.firewall.allowedTCPPorts`
- `explain-option programs.zsh.enable`

**Usage in TUI**:
1. Select `explain-option [INPUT]` from the command list
2. Tab to execution panel
3. Add the option path after `explain-option`
4. Press Enter to execute

**What you get**:
- Detailed option explanation
- Valid values and types
- Configuration examples
- Related options and dependencies

### 4. `package-repo [INPUT]` - Repository Analysis
**Purpose**: Analyze Git repositories and generate Nix derivations

**Examples**:
- `package-repo https://github.com/user/project`
- `package-repo https://gitlab.com/org/repo.git`
- `package-repo git@github.com:user/private-repo.git`

**Usage in TUI**:
1. Select `package-repo [INPUT]` from the command list
2. Tab to execution panel
3. Add the repository URL after `package-repo`
4. Press Enter to execute

**What you get**:
- Automatic dependency detection
- Generated Nix derivation
- Build instructions and configuration
- Integration suggestions

## How to Use Input Commands in TUI

### Method 1: Command Selection + Input Entry
```
1. Navigate to command list (left panel)
2. Use ↑↓ arrow keys to select command with [INPUT]
3. Press Enter → command appears in execution panel
4. Type your input after the command name
5. Press Enter to execute
```

### Method 2: Direct Typing in Execution Panel
```
1. Press Tab to switch to execution panel (right side)
2. Type complete command with input:
   - ask "your question here"
   - search packagename
   - explain-option option.path
   - package-repo https://repo-url
3. Press Enter to execute
```

### Method 3: Search + Select
```
1. Press / to open search mode
2. Type part of command name (e.g., "ask", "search")
3. Select from filtered results
4. Follow Method 1 steps
```

## Input Command Features

### Smart Quoting
- **Questions**: Use quotes for multi-word questions: `ask "how to setup SSH?"`
- **Search terms**: Quotes optional but recommended for phrases: `search "text editor"`
- **Options**: No quotes needed: `explain-option services.ssh.enable`
- **URLs**: No quotes needed: `package-repo https://github.com/user/repo`

### Input History
- Use ↑↓ arrows in execution panel to navigate command history
- Previously executed commands are remembered
- Easy to re-run with modifications

### Auto-completion Hints
- Commands show usage examples in the left panel
- Input field shows placeholder text
- Error messages provide correction suggestions

### Real-time Feedback
- Command execution shows progress indicators
- Streaming output for long-running operations
- Clear success/error status indicators

## Tips for Effective Usage

### For `ask` Commands:
- Be specific in your questions
- Include context: "as a beginner" or "for gaming setup"
- Ask follow-up questions for clarification
- Use examples: "show me a configuration example"

### For `search` Commands:
- Use descriptive terms: "web browser" vs just "browser"
- Try different synonyms if first search doesn't find what you need
- Search by category: "development tools", "graphics drivers"

### For `explain-option` Commands:
- Use tab completion in the command line for option paths
- Start with broader options and drill down: `services` → `services.nginx`
- Look at the examples provided in command output

### For `package-repo` Commands:
- Ensure repository is publicly accessible
- Use HTTPS URLs when possible
- Repository should have clear build instructions
- Check for existing Nix derivations first

## Integration with Other Features

### With Documentation Search:
- Input commands automatically query NixOS documentation
- Results include official manual references
- Links to community resources and examples

### With AI Providers:
- All input commands leverage configured AI provider
- Responses tailored to your experience level
- Context-aware suggestions based on your system

### With Configuration Management:
- Results can be directly applied to your configuration
- Generated code snippets are NixOS-ready
- Integration with templates and snippets system

## Troubleshooting Input Commands

### Common Issues:
1. **Command not found**: Ensure proper spelling and syntax
2. **No results**: Try broader or alternative search terms
3. **Permission errors**: Check repository access permissions
4. **Network timeouts**: Verify internet connectivity

### Getting Help:
- Press `F1` for changelog and recent features
- Use `ask "help with [specific command]"` for command-specific help
- Check `nixai --help` for CLI documentation

## Keyboard Shortcuts Summary

| Action | Shortcut | Context |
|--------|----------|---------|
| Switch panels | `Tab` | Any panel |
| Navigate commands | `↑↓` | Command list |
| Search commands | `/` | Command list |
| Navigate history | `↑↓` | Execution panel (when focused) |
| Execute command | `Enter` | Execution panel |
| Exit TUI | `Ctrl+C` or `q` | Any panel |
| Show changelog | `F1` | Any panel |

---

*This guide covers the interactive TUI usage. For CLI usage examples, see the main documentation or run `nixai [command] --help`.*
