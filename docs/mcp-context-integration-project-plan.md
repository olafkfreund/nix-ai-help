# 🔗 MCP Context Integration Project Plan

**Project**: Expose nixai context command functionality through MCP server for VS Code and Neovim integration

**Created**: 10 June 2025  
**Status**: 🚧 Planning Phase  
**Priority**: High  

---

## 📋 Project Overview

### **Objective**
Integrate the `nixai context` command functionality into the MCP (Model Context Protocol) server to make NixOS system context information available to AI assistants in VS Code and Neovim. This will enable context-aware AI responses that adapt to the user's actual NixOS configuration setup.

### **Current State**
- ✅ **Context System**: Robust context detection system with 4 subcommands (`detect`, `show`, `reset`, `status`)
- ✅ **MCP Server**: Fully functional with 4 existing tools (query_nixos_docs, explain_nixos_option, etc.)
- ✅ **VS Code Integration**: Working MCP integration via Unix sockets
- ✅ **Neovim Integration**: Basic MCP integration available

### **Target Benefits**
- **Context-Aware AI**: AI assistants provide responses tailored to actual NixOS setup (flakes vs channels, Home Manager type, etc.)
- **Smart Configuration Suggestions**: Responses automatically adapt to user's configuration type
- **Real-time Context Access**: Get system context without leaving the editor
- **Enhanced Developer Experience**: Seamless integration between CLI context and editor AI assistance

---

## 🎯 Implementation Phases

### **Phase 1: Core MCP Tools Implementation** ⏳
**Duration**: 2-3 days  
**Priority**: Critical  

#### Tasks
- [ ] Add 4 new MCP tools to server tool list:
  - `get_nixos_context` - Get current system context
  - `detect_nixos_context` - Force context re-detection  
  - `reset_nixos_context` - Clear cache and refresh
  - `context_status` - Show context system health

- [ ] Implement handler functions in MCP server
- [ ] Create context formatting utilities for MCP responses
- [ ] Add error handling and validation
- [ ] Write unit tests for new MCP tools

#### Files to Modify
- `internal/mcp/server.go` - Add tools and handlers
- `internal/mcp/context_handlers.go` - New file for context-specific handlers
- `internal/mcp/context_formatters.go` - New file for response formatting

---

### **Phase 2: Enhanced Context Response Formatting** ⏳  
**Duration**: 1-2 days  
**Priority**: High  

#### Tasks
- [ ] Create structured JSON response format for context data
- [ ] Add human-readable text formatting option
- [ ] Implement context diff detection (show what changed)
- [ ] Add context validation and health checks via MCP
- [ ] Create context summary for AI prompt integration

#### Files to Modify
- `internal/mcp/context_formatters.go` - Response formatting logic
- `internal/config/context.go` - Context data structures if needed

---

### **Phase 3: VS Code Integration Enhancement** ⏳
**Duration**: 1-2 days  
**Priority**: High  

#### Tasks
- [ ] Update VS Code MCP configuration documentation
- [ ] Add context tools to MCP server settings examples
- [ ] Create VS Code-specific usage examples
- [ ] Test integration with Copilot and Claude Dev
- [ ] Add context information to AI prompt templates

#### Files to Modify
- `docs/MCP_VSCODE_INTEGRATION.md` - Updated documentation
- `modules/home-manager.nix` - VS Code integration settings
- `scripts/mcp-bridge.sh` - If needed for VS Code connection

---

### **Phase 4: Neovim Integration Enhancement** ⏳
**Duration**: 1-2 days  
**Priority**: Medium  

#### Tasks
- [ ] Update Neovim integration to use context MCP tools
- [ ] Add Lua functions for context access
- [ ] Create Neovim commands for context management
- [ ] Add context display in status line/floating windows
- [ ] Update Neovim documentation

#### Files to Modify
- `internal/neovim/integration.go` - Enhanced Neovim functions
- `docs/neovim-integration.md` - Updated documentation
- `modules/nixvim-nixai-example.nix` - Example configuration

---

### **Phase 5: Documentation and Testing** ⏳
**Duration**: 1 day  
**Priority**: Medium  

#### Tasks
- [ ] Create comprehensive usage examples
- [ ] Add context integration to README
- [ ] Write integration tests for MCP context tools
- [ ] Update MCP server documentation
- [ ] Create troubleshooting guide

#### Files to Modify
- `README.md` - Updated feature list
- `docs/mcp-server.md` - Updated documentation
- `tests/mcp/test_context_integration.py` - New test file
- `docs/MANUAL.md` - Updated command documentation

---

## 🔧 Technical Implementation Details

### **New MCP Tools Specification**

#### 1. `get_nixos_context`
```json
{
  "name": "get_nixos_context",
  "description": "Get current NixOS system context information",
  "inputSchema": {
    "type": "object",
    "properties": {
      "format": {"type": "string", "enum": ["json", "text"], "default": "json"},
      "detailed": {"type": "boolean", "default": false}
    }
  }
}
```

#### 2. `detect_nixos_context`
```json
{
  "name": "detect_nixos_context", 
  "description": "Force re-detection of NixOS system context",
  "inputSchema": {
    "type": "object",
    "properties": {
      "verbose": {"type": "boolean", "default": false}
    }
  }
}
```

#### 3. `reset_nixos_context`
```json
{
  "name": "reset_nixos_context",
  "description": "Clear cached context and force refresh",
  "inputSchema": {
    "type": "object",
    "properties": {
      "confirm": {"type": "boolean", "default": true}
    }
  }
}
```

#### 4. `context_status`
```json
{
  "name": "context_status",
  "description": "Show context detection system status and health",
  "inputSchema": {
    "type": "object",
    "properties": {
      "include_metrics": {"type": "boolean", "default": false}
    }
  }
}
```

### **Response Format Examples**

#### Context Response Format
```json
{
  "content": [{
    "type": "text",
    "text": "📋 System: nixos | Flakes: Yes | Home Manager: standalone\n\n### System Information\n..."
  }],
  "context": {
    "systemType": "nixos",
    "usesFlakes": true,
    "homeManagerType": "standalone",
    "nixosVersion": "25.11.20250607.3e3afe5",
    "enabledServices": ["nginx", "postgresql", "ssh"],
    "configPaths": {
      "nixos": "/etc/nixos",
      "flake": "/etc/nixos/flake.nix"
    },
    "cacheInfo": {
      "valid": true,
      "lastDetected": "2025-06-10T18:56:02Z",
      "ageSeconds": 135
    }
  }
}
```

### **Integration Architecture**

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   VS Code       │    │   MCP Server     │    │ Context System  │
│   (Copilot/     │◄──►│   (Unix Socket)  │◄──►│ (nixai context) │
│    Claude)      │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
┌─────────────────┐             │
│   Neovim        │             │
│   (nixai.nvim)  │◄────────────┘
└─────────────────┘
```

---

## 🧪 Testing Strategy

### **Unit Tests**
- [ ] Test each MCP tool individually
- [ ] Test context formatting functions
- [ ] Test error handling scenarios
- [ ] Test cache validation logic

### **Integration Tests** 
- [ ] Test MCP server with new context tools
- [ ] Test VS Code integration end-to-end
- [ ] Test Neovim integration functionality
- [ ] Test context updates and cache invalidation

### **Manual Testing**
- [ ] Verify context tools work in VS Code with Copilot
- [ ] Verify context tools work with Claude Dev
- [ ] Test Neovim integration with context tools
- [ ] Test context-aware AI responses

---

## 📊 Success Metrics

### **Functional Requirements**
- [ ] All 4 context MCP tools implemented and working
- [ ] Context information accessible from VS Code AI assistants
- [ ] Context information accessible from Neovim
- [ ] AI responses adapt based on detected context
- [ ] Performance: Context retrieval < 500ms

### **User Experience Requirements**
- [ ] Context integration feels seamless and natural
- [ ] Documentation is clear and comprehensive
- [ ] Error messages are helpful and actionable
- [ ] Configuration is simple and reliable

---

## 🚀 Usage Examples (Target State)

### **VS Code with Copilot**
```
User: "How do I enable SSH?"

Copilot (with context): "Since you're using flakes and have Home Manager 
as a standalone setup, here's the configuration:

For NixOS (system-level):
```nix
# In your flake.nix
services.openssh = {
  enable = true;
  settings.PasswordAuthentication = false;
};
```

For Home Manager (user-level):
```nix  
# In your home.nix
programs.ssh = {
  enable = true;
  # ... user SSH configuration
};
```
```

### **Neovim Integration**
```lua
-- Get current context
:lua local ctx = require('nixai-nvim').get_context()
-- Ask context-aware question  
:NixaiAsk "How do I configure this service?"
-- AI response will be tailored to your setup
```

---

## 📅 Timeline

| Phase | Duration | Start Date | End Date | Status |
|-------|----------|------------|----------|---------|
| Phase 1 | 2-3 days | TBD | TBD | ⏳ Pending |
| Phase 2 | 1-2 days | TBD | TBD | ⏳ Pending |
| Phase 3 | 1-2 days | TBD | TBD | ⏳ Pending |
| Phase 4 | 1-2 days | TBD | TBD | ⏳ Pending |
| Phase 5 | 1 day | TBD | TBD | ⏳ Pending |
| **Total** | **7-10 days** | TBD | TBD | ⏳ Planning |

---

## 🔄 Progress Tracking

### **Completed Tasks** ✅
- [x] Project planning and specification
- [x] Technical architecture design
- [x] Documentation structure planned

### **In Progress Tasks** 🚧
- [ ] *Ready to start Phase 1*

### **Blocked/Issues** 🚫
- None currently identified

---

## 📝 Notes and Considerations

### **Technical Considerations**
- Reuse existing context detection logic from `internal/cli/context_commands.go`
- Maintain consistency with existing MCP tool patterns
- Ensure proper error handling for context detection failures
- Consider performance impact of frequent context queries

### **User Experience Considerations**
- Context should be transparent to users (work automatically)
- Provide option to manually refresh context when needed
- Clear error messages when context detection fails
- Graceful degradation when context unavailable

### **Future Enhancements** 🔮
- Context change notifications/webhooks
- Multiple context profiles (dev/prod/test)
- Context-based configuration templates
- Integration with other editors (Emacs, Vim, etc.)

---

**Project Lead**: AI Assistant  
**Reviewer**: User  
**Last Updated**: 10 June 2025
