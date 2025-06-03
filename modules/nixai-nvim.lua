-- nixai-nvim.lua: Integration with nixai MCP server (multi-endpoint)
local M = {}

local api = vim.api

-- Set endpoints and default socket path from global variables (populated by Nix)
M.endpoints = vim.g.nixai_endpoints or {}
M.default_socket_path = vim.g.nixai_socket_path or os.getenv("HOME") .. "/.local/share/nixai/mcp.sock"

-- nvim-notify integration (feature 7)
local function notify(msg, level)
  local ok, notify = pcall(require, 'notify')
  if ok then
    notify(msg, level or vim.log.levels.INFO, { title = 'nixai' })
  else
    vim.schedule(function() vim.api.nvim_echo({{msg}}, true, {}) end)
  end
end

-- Floating window helper
local function open_floating_markdown(lines, title)
  local buf = api.nvim_create_buf(false, true)
  api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  local width = math.max(60, math.floor(vim.o.columns * 0.6))
  local height = math.max(10, #lines + 2)
  local win = api.nvim_open_win(buf, true, {
    relative = 'editor',
    width = width,
    height = height,
    row = math.floor((vim.o.lines - height) / 2),
    col = math.floor((vim.o.columns - width) / 2),
    style = 'minimal',
    border = 'rounded',
    title = title or 'nixai',
  })
  api.nvim_buf_set_keymap(buf, 'n', 'q', '<cmd>close<CR>', { nowait = true, noremap = true, silent = true })
  return buf, win
end

-- Improved query with error handling and floating output
function M.query(question, endpoint)
  local socket_path = M.default_socket_path
  if endpoint then
    for _, ep in ipairs(M.endpoints) do
      if ep.name == endpoint then
        socket_path = ep.socketPath
        break
      end
    end
  end
  local cmd = string.format("nixai --ask '%s' --socket-path=%s", question:gsub("'", "''"), socket_path)
  local output = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    open_floating_markdown({"**nixai error:**", output}, "nixai error")
    return
  end
  local lines = vim.split(output, "\n")
  open_floating_markdown(lines, "nixai response")
end

-- Endpoint picker UI
function M.pick_endpoint_and_query(question)
  if #M.endpoints == 0 then
    M.query(question)
    return
  end
  local names = {}
  for _, ep in ipairs(M.endpoints) do table.insert(names, ep.name) end
  vim.ui.select(names, { prompt = 'Select nixai endpoint:' }, function(choice)
    if choice then M.query(question, choice) end
  end)
end

-- Telescope integration for fuzzy doc search
function M.telescope_search(endpoint)
  local has_telescope, telescope = pcall(require, 'telescope.builtin')
  if not has_telescope then
    open_floating_markdown({
      "# nixai Telescope integration",
      "",
      "❌ Telescope.nvim not found!",
      "Install telescope.nvim to use fuzzy doc search.",
    }, "nixai error")
    return
  end
  local socket_path = M.default_socket_path
  if endpoint then
    for _, ep in ipairs(M.endpoints) do
      if ep.name == endpoint then
        socket_path = ep.socketPath
        break
      end
    end
  end
  -- Query all options/docs from MCP (assume nixai --list-options returns JSON list)
  local cmd = string.format("nixai mcp-list-options --socket-path=%s", socket_path)
  local output = vim.fn.system(cmd)
  if vim.v.shell_error ~= 0 then
    open_floating_markdown({"**nixai error:**", output}, "nixai error")
    return
  end
  local ok, opts = pcall(vim.fn.json_decode, output)
  if not ok or type(opts) ~= "table" then
    open_floating_markdown({"**nixai error:**", "Could not parse MCP options output."}, "nixai error")
    return
  end
  telescope.new({
    prompt_title = "NixOS/Home Manager Options",
    finder = telescope.finders.new_table {
      results = opts,
      entry_maker = function(entry)
        return {value = entry, display = entry, ordinal = entry}
      end,
    },
    sorter = telescope.config.values.generic_sorter({}),
    attach_mappings = function(_, map)
      map('i', '<CR>', function(prompt_bufnr)
        local selection = require('telescope.actions.state').get_selected_entry()
        require('telescope.actions').close(prompt_bufnr)
        if selection then
          M.query('Explain option: ' .. selection.value, endpoint)
        end
      end)
      return true
    end,
  }):find()
end

-- Show current MCP endpoint in statusline (feature 7)
function M.statusline_endpoint()
  if #M.endpoints == 0 then
    return '[nixai:default]'
  end
  local current = M.current_endpoint or M.endpoints[1].name
  return string.format('[nixai:%s]', current)
end

-- NixaiHelp: show keymaps and commands
function M.nixai_help()
  local lines = {
    "# Nixai Neovim Integration Help",
    "",
    "**Keymaps:**",
    "- <leader>na: Ask nixai (default endpoint)",
    "- <leader>nX: Ask nixai for endpoint X (e.g. <leader>nd for 'default')",
    "- Visual + <leader>na: Ask nixai about selection",
    "",
    "**Commands:**",
    ":NixaiHelp - Show this help",
    ":NixaiDoctor - Run diagnostics",
    ":NixaiInstallDeps - Show install instructions for missing dependencies",
    ":NixaiUpdate - Auto-update/install nixai and MCP",
    ":lua require'nixai-nvim'.telescope_search() - Fuzzy doc search (requires telescope.nvim)",
    "",
    "**Tips:**",
    "- Use endpoint picker: :lua require'nixai-nvim'.pick_endpoint_and_query('Your question')",
    "- Configure endpoints via Nix module or vim.g.nixai_endpoints",
  }
  open_floating_markdown(lines, "Nixai Help")
end

-- NixaiDoctor: check endpoints and dependencies
function M.nixai_doctor()
  local lines = {"# Nixai Doctor", ""}
  -- Check dependencies
  local function check_dep(dep)
    return (vim.fn.executable(dep) == 1) and ("✅ " .. dep .. " found") or ("❌ " .. dep .. " missing")
  end
  table.insert(lines, "## Dependencies:")
  for _, dep in ipairs({"nixai", "socat"}) do
    table.insert(lines, "- " .. check_dep(dep))
  end
  table.insert(lines, "")
  -- Check endpoints
  table.insert(lines, "## MCP Endpoints:")
  if #M.endpoints == 0 then
    table.insert(lines, "- (none configured, using default socket)")
  else
    for _, ep in ipairs(M.endpoints) do
      local ok = (vim.fn.filereadable(ep.socketPath) == 1) and "✅" or "❌"
      table.insert(lines, string.format("- %s [%s:%d] %s", ep.name, ep.host, ep.port, ok))
    end
  end
  open_floating_markdown(lines, "Nixai Doctor")
end

-- Auto-install/update nixai and MCP (feature 8)
function M.nixai_update()
  local lines = { '# nixai auto-update', '' }
  local update_cmd = 'nix-env -iA nixpkgs/nixai'
  local mcp_cmd = 'nix-env -iA nixpkgs/socat'
  table.insert(lines, 'Updating nixai and socat using:')
  table.insert(lines, '')
  table.insert(lines, '  ' .. update_cmd)
  table.insert(lines, '  ' .. mcp_cmd)
  table.insert(lines, '')
  table.insert(lines, 'This will run in your shell. Proceed? (y/n)')
  open_floating_markdown(lines, 'nixai update')
  vim.ui.input({ prompt = 'Run update commands? (y/n): ' }, function(input)
    if input and input:lower() == 'y' then
      notify('Updating nixai and socat...', vim.log.levels.INFO)
      vim.fn.jobstart(update_cmd .. ' && ' .. mcp_cmd, {
        on_stdout = function(_, data)
          if data then notify(table.concat(data, '\n'), vim.log.levels.INFO) end
        end,
        on_stderr = function(_, data)
          if data then notify(table.concat(data, '\n'), vim.log.levels.ERROR) end
        end,
        on_exit = function(_, code)
          if code == 0 then
            notify('nixai and socat updated!', vim.log.levels.INFO)
          else
            notify('Update failed. See output above.', vim.log.levels.ERROR)
          end
        end,
      })
    else
      notify('Update cancelled.', vim.log.levels.WARN)
    end
  end)
end

-- Install dependencies (show instructions for zsh)
function M.nixai_install_deps()
  local missing = {}
  for _, dep in ipairs({"nixai", "socat"}) do
    if vim.fn.executable(dep) ~= 1 then table.insert(missing, dep) end
  end
  if #missing == 0 then
    open_floating_markdown({"All dependencies are installed!"}, "nixai deps")
    return
  end
  local lines = {"# nixai missing dependencies:", ""}
  for _, dep in ipairs(missing) do
    table.insert(lines, "- " .. dep)
  end
  table.insert(lines, "", "Install with:")
  table.insert(lines, "zsh:")
  table.insert(lines, "  nix-env -iA nixpkgs/" .. table.concat(missing, " nixpkgs/"))
  open_floating_markdown(lines, "nixai deps")
end

-- Setup keymaps for each endpoint
function M.setup_keymaps()
  for _, ep in ipairs(M.endpoints) do
    vim.keymap.set("n", "<leader>n" .. string.sub(ep.name, 1, 1), function()
      M.query("Ask nixai (" .. ep.name .. ")", ep.name)
    end, { desc = "Ask nixai (" .. ep.name .. ")" })
  end
  -- Default keymaps
  vim.keymap.set("n", "<leader>na", function() M.query("Ask nixai") end, { desc = "Ask nixai" })
  vim.keymap.set("v", "<leader>na", function()
    local start_pos = vim.fn.getpos("'<")
    local end_pos = vim.fn.getpos("'>")
    local lines = vim.fn.getline(start_pos[2], end_pos[2])
    local text = table.concat(lines, "\n")
    M.query("Explain this code: " .. text)
  end, { desc = "Ask nixai about selection" })
  -- Add commands
  vim.api.nvim_create_user_command('NixaiHelp', M.nixai_help, {})
  vim.api.nvim_create_user_command('NixaiDoctor', M.nixai_doctor, {})
  vim.api.nvim_create_user_command('NixaiInstallDeps', M.nixai_install_deps, {})
  vim.api.nvim_create_user_command('NixaiUpdate', M.nixai_update, {})
end

print("nixai integration loaded! Use <leader>nX for each endpoint, or <leader>na for default. :NixaiHelp for help.")

return M
