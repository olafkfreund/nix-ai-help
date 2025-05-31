# Neovim Integration for nixai

This document explains how to integrate nixai with Neovim for seamless NixOS documentation and assistance directly in your editor.

## Setup

### Automatic Setup

Use the built-in command to automatically set up Neovim integration:

```bash
nixai neovim-setup
```

This will:
1. Create a `nixai.lua` module in your Neovim configuration
2. Provide instructions for adding it to your `init.lua` or `init.vim`

#### Options:

- `--socket-path`: Specify a custom socket path (default: /tmp/nixai-mcp.sock)
- `--config-dir`: Specify a custom Neovim config directory (default: auto-detected)

Example:
```bash
nixai neovim-setup --socket-path=$HOME/.local/share/nixai/mcp.sock
```

### Manual Setup

1. Create a file at `~/.config/nvim/lua/nixai.lua` with the code shown in [Manual Module Setup](#manual-module-setup)

2. Add to your `init.lua`:
```lua
-- nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "/tmp/nixai-mcp.sock",  -- Adjust to match your MCP socket path
  })
else
  vim.notify("nixai module not found", vim.log.levels.WARN)
end
```

Or if you're using `init.vim`:
```vim
lua << EOF
-- nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "/tmp/nixai-mcp.sock",  -- Adjust to match your MCP socket path
  })
else
  vim.notify("nixai module not found", vim.log.levels.WARN)
end
EOF
```

## Usage

### Default Key Mappings

- `<leader>nq` - Ask a NixOS question
- `<leader>ns` - Get context-aware NixOS suggestions
- `<leader>no` - Explain a NixOS option
- `<leader>nh` - Explain a Home Manager option

### Custom Usage

You can call nixai functions directly in your Lua code:

```lua
-- Query NixOS documentation
require('nixai').query_docs("How do I configure nginx?")

-- Explain a NixOS option
require('nixai').explain_option("services.nginx.enable")

-- Explain a Home Manager option
require('nixai').explain_home_option("programs.git.enable")

-- Search NixOS packages
require('nixai').search_packages("firefox")
```

### Integration with Telescope

Create a custom Telescope picker for nixai:

```lua
local nixai_picker = function()
  local pickers = require('telescope.pickers')
  local finders = require('telescope.finders')
  local actions = require('telescope.actions')
  local action_state = require('telescope.actions.state')
  local conf = require('telescope.config').values
  
  local nixai = require('nixai')
  
  pickers.new({}, {
    prompt_title = 'NixOS Query',
    finder = finders.new_dynamic({
      fn = function(prompt)
        if prompt and #prompt > 0 then
          local result = nixai.query_docs(prompt)
          if result and result.content and result.content[1] then
            return {{value = result.content[1].text, display = "NixOS: " .. prompt}}
          else
            return {{value = "No results found", display = "No results"}}
          end
        end
        return {}
      end,
      entry_maker = function(entry)
        return {
          value = entry,
          display = entry.display,
          ordinal = entry.display,
        }
      end,
    }),
    sorter = conf.generic_sorter({}),
    attach_mappings = function(prompt_bufnr, map)
      actions.select_default:replace(function()
        actions.close(prompt_bufnr)
        local selection = action_state.get_selected_entry()
        nixai.show_in_float({
          content = {{text = selection.value.value}}
        }, "NixOS: " .. selection.value.display)
      end)
      return true
    end,
  }):find()
end

-- Map to a key combination
vim.keymap.set('n', '<leader>nt', nixai_picker, {desc = 'Telescope NixOS Query'})
```

## Customizing the Setup

You can customize the nixai Neovim integration:

```lua
require('nixai').setup({
  socket_path = "/custom/path/to/mcp.sock",
  disable_keymaps = true,  -- Disable default keymaps if you want to define your own
})
```

Then define your own keymaps:

```lua
local nixai = require('nixai')

-- Custom keymaps
vim.keymap.set('n', '<leader>nx', nixai.ask_query, {desc = 'NixAI Query'})
vim.keymap.set('n', '<leader>nr', nixai.show_suggestion, {desc = 'NixAI Suggestion'})
```

## Dependencies

The nixai Neovim integration requires:

1. **socat**: For communication with the Unix socket
2. **Running MCP server**: Make sure the nixai MCP server is running

```bash
# Install socat
sudo nixos-rebuild switch --update-input nixpkgs --flake '.#nixos-config' -I nixos-config=./configuration.nix -I nixpkgs-overlays=./overlays --upgrade

# Start MCP server (if not already running)
nixai mcp-server start --background
```

## NixOS Integration

Add this to your NixOS/Home Manager configuration:

```nix
# Home Manager configuration
{ config, pkgs, ... }: {
  imports = [
    # Path to the nixai flake or local module
    (builtins.fetchTarball "https://github.com/olafkfreund/nix-ai-help/archive/main.tar.gz")/modules/home-manager.nix
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      # Optional: custom socket path (uses $HOME expansion)
      socketPath = "$HOME/.local/share/nixai/mcp.sock";
    };
  };
  
  programs.neovim = {
    enable = true;
    # Ensure socat is available
    extraPackages = with pkgs; [ socat ];
    
    # Add the nixai integration to your Neovim config
    extraLuaConfig = ''
      -- nixai integration
      local ok, nixai = pcall(require, "nixai")
      if ok then
        nixai.setup({
          socket_path = "${config.home.homeDirectory}/.local/share/nixai/mcp.sock",
        })
      end
    '';
  };
}
```

## Manual Module Setup

<details>
<summary>Click to expand nixai.lua module code</summary>

```lua
-- nixai.lua: Integration with nixai MCP server
local M = {}

-- Socket path for the nixai MCP server
M.socket_path = "/tmp/nixai-mcp.sock"

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
  end
end

return M
```
</details>

## Troubleshooting

### Common Issues

1. **"Failed to connect to socket"** - Make sure the MCP server is running
   ```bash
   nixai mcp-server status
   # If not running:
   nixai mcp-server start
   ```

2. **"socat: command not found"** - Install socat
   ```bash
   nix-env -iA nixos.socat
   ```

3. **"Failed to parse JSON response"** - Check that the MCP server is responding correctly
   ```bash
   socat - UNIX-CONNECT:/tmp/nixai-mcp.sock
   # And then try a simple request:
   {"jsonrpc":"2.0","id":1,"method":"initialize"}
   ```

4. **Wrong socket path** - Update the path in your Neovim config
   ```lua
   require('nixai').setup({
     socket_path = "/correct/path/to/mcp.sock",
   })
   ```

5. **Module not found** - Make sure the module is correctly installed
   ```bash
   # Verify the module exists:
   ls ~/.config/nvim/lua/nixai.lua
   
   # Or run the setup again:
   nixai neovim-setup
   ```
