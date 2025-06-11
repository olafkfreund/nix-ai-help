// Package neovim provides integration between nixai's MCP server and Neovim
package neovim

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ConfigData holds data for generating Neovim configuration
type ConfigData struct {
	SocketPath string
}

// GetUserConfigDir returns the user's Neovim config directory
func GetUserConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check for different Neovim config paths
	configPaths := []string{
		filepath.Join(homeDir, ".config", "nvim"),
		filepath.Join(homeDir, ".nvim"),
		filepath.Join(homeDir, "AppData", "Local", "nvim"), // Windows
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Default to ~/.config/nvim if no existing config found
	return filepath.Join(homeDir, ".config", "nvim"), nil
}

// CreateNeovimModule creates the nixai integration module for Neovim
func CreateNeovimModule(socketPath string, configDir string) error {
	if configDir == "" {
		var err error
		configDir, err = GetUserConfigDir()
		if err != nil {
			return err
		}
	}

	// Create lua directory if it doesn't exist
	luaDir := filepath.Join(configDir, "lua")
	if err := os.MkdirAll(luaDir, 0755); err != nil {
		return fmt.Errorf("failed to create lua directory: %w", err)
	}

	// Create the nixai.lua file
	nixaiLuaPath := filepath.Join(luaDir, "nixai.lua")
	nixaiLuaFile, err := os.Create(nixaiLuaPath)
	if err != nil {
		return fmt.Errorf("failed to create nixai.lua: %w", err)
	}
	defer func() { _ = nixaiLuaFile.Close() }()

	// Use default socket path if not provided
	if socketPath == "" {
		socketPath = "/tmp/nixai-mcp.sock"
	}

	// Replace $HOME with the literal string "$HOME" for the template
	socketPath = strings.ReplaceAll(socketPath, os.Getenv("HOME"), "$HOME")

	data := ConfigData{
		SocketPath: socketPath,
	}

	// Parse the template
	tmpl, err := template.New("nixai.lua").Parse(neovimModuleTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template
	if err := tmpl.Execute(nixaiLuaFile, data); err != nil {
		return fmt.Errorf("failed to generate nixai.lua: %w", err)
	}

	return nil
}

// GenerateInitConfig returns a Lua snippet to add to init.lua
func GenerateInitConfig(socketPath string) string {
	if socketPath == "" {
		socketPath = "/tmp/nixai-mcp.sock"
	}

	// Replace $HOME with the actual home directory for the snippet
	socketPath = strings.ReplaceAll(socketPath, "$HOME", os.Getenv("HOME"))

	return fmt.Sprintf(initLuaSnippet, socketPath)
}

// neovimModuleTemplate is the template for the nixai.lua file
const neovimModuleTemplate = `-- nixai.lua: Integration with nixai MCP server
local M = {}

-- Socket path for the nixai MCP server
M.socket_path = "{{.SocketPath}}"

-- Context buffer for current file
local function get_current_buffer_content()
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  return table.concat(lines, "\n")
end

-- Get current line context
local function get_current_context()
  local line = vim.api.nvim_get_current_line()
  local row, col = unpack(vim.api.nvim_win_get_cursor(0))
  local before_cursor = line:sub(1, col)
  
  -- Get a few lines before for context
  local start_row = math.max(0, row - 5)
  local context_lines = vim.api.nvim_buf_get_lines(0, start_row, row, false)
  table.insert(context_lines, before_cursor)
  
  return table.concat(context_lines, "\n")
end

-- Call the MCP server using socat
function M.call_mcp(tool, args)
  -- Create a temporary file for input
  local input_file = os.tmpname()
  local output_file = os.tmpname()
  
  -- Prepare input JSON for the MCP server
  local json = string.format([[
  {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "%s",
      "arguments": %s
    }
  }
  ]], tool, vim.json.encode(args))
  
  -- Write JSON to input file
  local f = io.open(input_file, "w")
  f:write(json)
  f:close()
  
  -- Call socat to communicate with the socket
  local cmd = string.format("cat %s | socat - UNIX-CONNECT:%s > %s", 
                            input_file, M.socket_path, output_file)
  os.execute(cmd)
  
  -- Read the response
  local f = io.open(output_file, "r")
  if not f then
    os.remove(input_file)
    os.remove(output_file)
    return nil, "Failed to open output file"
  end
  
  local response = f:read("*all")
  f:close()
  
  -- Clean up temporary files
  os.remove(input_file)
  os.remove(output_file)
  
  -- Parse the response
  local success, result = pcall(vim.json.decode, response)
  if not success then
    return nil, "Failed to parse JSON response: " .. response
  end
  
  if result.result then
    return result.result, nil
  else
    return nil, result.error and result.error.message or "Unknown error"
  end
end

-- Query NixOS documentation
function M.query_docs(query)
  local result, err = M.call_mcp("query_nixos_docs", {query = query})
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Explain NixOS option
function M.explain_option(option)
  local result, err = M.call_mcp("explain_nixos_option", {option = option})
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Explain Home Manager option
function M.explain_home_option(option)
  local result, err = M.call_mcp("explain_home_manager_option", {option = option})
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Search NixOS packages
function M.search_packages(query)
  local result, err = M.call_mcp("search_nixos_packages", {query = query})
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Show result in a floating window
function M.show_in_float(result, title)
  if not result or not result.content or not result.content[1] then
    vim.notify("No results from NixAI", vim.log.levels.WARN)
    return
  end
  
  local text = result.content[1].text
  
  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, true, vim.split(text, '\n'))
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  
  local width = math.min(80, vim.o.columns - 4)
  local height = math.min(20, vim.o.lines - 4)
  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)
  
  local opts = {
    relative = 'editor',
    width = width,
    height = height,
    row = row,
    col = col,
    style = 'minimal',
    border = 'rounded',
    title = title or 'NixAI',
    title_pos = 'center'
  }
  
  local win = vim.api.nvim_open_win(buf, true, opts)
  
  -- Close on q or ESC
  vim.keymap.set('n', 'q', function() vim.api.nvim_win_close(win, true) end, 
                {buffer = buf, noremap = true})
  vim.keymap.set('n', '<Esc>', function() vim.api.nvim_win_close(win, true) end, 
                {buffer = buf, noremap = true})
end

-- Get suggestion based on current context
function M.get_suggestion()
  local context = get_current_context()
  local filetype = vim.bo.filetype
  
  -- Use appropriate MCP tool based on filetype
  local tool = "query_nixos_docs"
  local args = {}
  
  if filetype == "nix" then
    -- Look for specific Nix patterns
    if context:match("services%.") then
      tool = "explain_nixos_option"
      args.option = context:match("services%.[%w%-]+")
    elseif context:match("programs%.") and context:match("home%-manager") then
      tool = "explain_home_manager_option"
      args.option = context:match("programs%.[%w%-]+")
    else
      args.query = "Context: " .. context .. "\nWhat Nix configuration would help here?"
    end
  else
    args.query = "Context: " .. context .. "\nWhat Nix configuration would help here?"
  end
  
  local result, err = M.call_mcp(tool, args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  
  return result
end

-- Show suggestion in a floating window
function M.show_suggestion()
  local suggestion = M.get_suggestion()
  if suggestion then
    M.show_in_float(suggestion, 'NixAI Suggestion')
  end
end

-- Ask query and show result
function M.ask_query()
  vim.ui.input({prompt = "NixOS Question: "}, function(input)
    if input and input ~= "" then
      local result = M.query_docs(input)
      if result then
        M.show_in_float(result, 'NixAI: ' .. input)
      end
    end
  end)
end

-- Get NixOS system context
function M.get_context(format, detailed)
  local args = {}
  if format then args.format = format end
  if detailed then args.detailed = detailed end
  
  local result, err = M.call_mcp("get_nixos_context", args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Force context re-detection
function M.detect_context(verbose)
  local args = {}
  if verbose then args.verbose = verbose end
  
  local result, err = M.call_mcp("detect_nixos_context", args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Reset context cache
function M.reset_context(confirm)
  local args = {}
  if confirm ~= nil then args.confirm = confirm end
  
  local result, err = M.call_mcp("reset_nixos_context", args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Get context system status
function M.context_status(include_metrics)
  local args = {}
  if include_metrics then args.includeMetrics = include_metrics end
  
  local result, err = M.call_mcp("context_status", args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Get context differences
function M.context_diff()
  local result, err = M.call_mcp("context_diff", {})
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  return result
end

-- Show current context in floating window
function M.show_context()
  local context = M.get_context("text", false)
  if context then
    M.show_in_float(context, 'NixOS System Context')
  end
end

-- Show detailed context in floating window
function M.show_detailed_context()
  local context = M.get_context("text", true)
  if context then
    M.show_in_float(context, 'NixOS System Context (Detailed)')
  end
end

-- Show context status in floating window
function M.show_context_status()
  local status = M.context_status(true)
  if status then
    M.show_in_float(status, 'NixOS Context System Status')
  end
end

-- Show context diff in floating window
function M.show_context_diff()
  local diff = M.context_diff()
  if diff then
    M.show_in_float(diff, 'NixOS Context Changes')
  end
end

-- Context-aware suggestion that uses system context
function M.get_context_aware_suggestion()
  local context = get_current_context()
  local filetype = vim.bo.filetype
  
  -- Get system context first
  local system_context = M.get_context("text", false)
  local context_info = ""
  if system_context and system_context.content and system_context.content[1] then
    -- Extract just the summary line from system context
    local lines = vim.split(system_context.content[1].text, '\n')
    for _, line in ipairs(lines) do
      if line:match("System Summary:") or line:match("📋 System:") then
        context_info = line .. " "
        break
      end
    end
  end
  
  -- Use appropriate MCP tool based on filetype and context
  local tool = "query_nixos_docs"
  local args = {}
  
  if filetype == "nix" then
    -- Look for specific Nix patterns
    if context:match("services%.") then
      tool = "explain_nixos_option"
      args.option = context:match("services%.[%w%-%.]+")
    elseif context:match("programs%.") and (context:match("home%-manager") or context_info:match("Home Manager")) then
      tool = "explain_home_manager_option"
      args.option = context:match("programs%.[%w%-%.]+")
    else
      args.query = context_info .. "Context: " .. context .. "\nWhat Nix configuration would help here?"
    end
  else
    args.query = context_info .. "Context: " .. context .. "\nWhat Nix configuration would help here?"
  end
  
  local result, err = M.call_mcp(tool, args)
  if err then
    vim.notify("NixAI Error: " .. err, vim.log.levels.ERROR)
    return nil
  end
  
  return result
end

-- Show context-aware suggestion in a floating window
function M.show_context_aware_suggestion()
  local suggestion = M.get_context_aware_suggestion()
  if suggestion then
    M.show_in_float(suggestion, 'NixAI Context-Aware Suggestion')
  end
end

-- Setup the nixai module
function M.setup(opts)
  opts = opts or {}
  if opts.socket_path then
    M.socket_path = opts.socket_path
  end
  
  -- Create namespace for virtual text
  if not vim.g.nixai_ns_id then
    vim.g.nixai_ns_id = vim.api.nvim_create_namespace("nixai_suggestions")
  end
  
  -- Setup keymaps
  if not opts.disable_keymaps then
    -- Original functionality
    vim.keymap.set('n', '<leader>ns', M.show_suggestion, {desc = 'NixAI Suggestion'})
    vim.keymap.set('n', '<leader>nq', M.ask_query, {desc = 'NixAI Query'})
    vim.keymap.set('n', '<leader>no', function()
      vim.ui.input({prompt = "NixOS Option: "}, function(input)
        if input and input ~= "" then
          local result = M.explain_option(input)
          if result then
            M.show_in_float(result, 'NixOS Option: ' .. input)
          end
        end
      end)
    end, {desc = 'NixAI Explain Option'})
    vim.keymap.set('n', '<leader>nh', function()
      vim.ui.input({prompt = "Home Manager Option: "}, function(input)
        if input and input ~= "" then
          local result = M.explain_home_option(input)
          if result then
            M.show_in_float(result, 'Home Manager Option: ' .. input)
          end
        end
      end)
    end, {desc = 'NixAI Explain Home Option'})
    
    -- Context-aware functionality ✨ NEW
    vim.keymap.set('n', '<leader>ncs', M.show_context_aware_suggestion, {desc = 'NixAI Context-Aware Suggestion'})
    vim.keymap.set('n', '<leader>ncc', M.show_context, {desc = 'NixAI Show Context'})
    vim.keymap.set('n', '<leader>ncd', M.show_detailed_context, {desc = 'NixAI Show Detailed Context'})
    vim.keymap.set('n', '<leader>ncr', function()
      vim.ui.select({'Yes', 'No'}, {prompt = 'Reset NixOS context cache?'}, function(choice)
        if choice == 'Yes' then
          local result = M.reset_context(true)
          if result then
            vim.notify("Context cache reset successfully", vim.log.levels.INFO)
            M.show_in_float(result, 'Context Reset')
          end
        end
      end)
    end, {desc = 'NixAI Reset Context'})
    vim.keymap.set('n', '<leader>nct', M.show_context_status, {desc = 'NixAI Context Status'})
    vim.keymap.set('n', '<leader>nck', M.show_context_diff, {desc = 'NixAI Context Changes'})
    vim.keymap.set('n', '<leader>ncf', function()
      local result = M.detect_context(true)
      if result then
        vim.notify("Context detection completed", vim.log.levels.INFO)
        M.show_in_float(result, 'Context Detection')
      end
    end, {desc = 'NixAI Force Context Detection'})
  end
end

return M
`

// initLuaSnippet is the snippet to add to init.lua
const initLuaSnippet = `
-- nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "%s",
  })
else
  vim.notify("nixai module not found", vim.log.levels.WARN)
end
`
