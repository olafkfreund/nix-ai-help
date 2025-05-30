#!/usr/bin/env bash
# docker-with-modules.sh
# Enhanced Docker setup script that utilizes nixai modules for proper Neovim and VS Code integration

set -euo pipefail

echo "🚀 Setting up nixai Docker environment with NixOS modules..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker first."
    exit 1
fi

# Build the Docker image
log_info "Building nixai Docker image..."
docker build -t nixai-nixos-modules -f docker_nixos/Dockerfile .

# Check if container already exists and stop it
if docker ps -a --format '{{.Names}}' | grep -q "nixai-modules-test"; then
    log_warning "Stopping existing container..."
    docker stop nixai-modules-test >/dev/null 2>&1 || true
    docker rm nixai-modules-test >/dev/null 2>&1 || true
fi

# Run the container with proper networking for Ollama
log_info "Starting nixai container with module support..."
docker run -it --name nixai-modules-test \
    --add-host=host.docker.internal:host-gateway \
    -e OLLAMA_HOST=http://host.docker.internal:11434 \
    nixai-nixos-modules bash -c "
        set -e
        echo '🐳 nixai Docker Environment with NixOS Modules'
        echo '=============================================='
        
        # Source Nix environment
        . /etc/profile.d/nix.sh
        
        cd /home/nixuser/nixai
        
        # Create Home Manager configuration with nixai module
        mkdir -p ~/.config/home-manager
        
        echo '🏠 Setting up Home Manager configuration with nixai module...'
        cat > ~/.config/home-manager/home.nix << 'EOF'
{ config, pkgs, ... }:

{
  imports = [ 
    ./nixai-module.nix
  ];

  # Allow unfree packages for Neovim plugins
  nixpkgs.config.allowUnfree = true;
  
  # Enable Neovim with extensive plugin support
  programs.neovim = {
    enable = true;
    defaultEditor = true;
    viAlias = true;
    vimAlias = true;
    
    plugins = with pkgs.vimPlugins; [
      # LSP and completion
      nvim-lspconfig
      nvim-cmp
      cmp-nvim-lsp
      cmp-buffer
      cmp-path
      luasnip
      
      # File management
      telescope-nvim
      nvim-tree-lua
      
      # Git integration
      gitsigns-nvim
      
      # UI enhancements
      lualine-nvim
      nvim-web-devicons
      
      # Nix language support
      vim-nix
      
      # AI/Chat integration (placeholder for future nixai plugin)
      # Will be replaced with actual nixai Neovim plugin
    ];
    
    extraConfig = '''
      \" Basic Neovim configuration
      set number relativenumber
      set expandtab tabstop=2 shiftwidth=2
      set hidden
      set ignorecase smartcase
      set termguicolors
      
      \" Enable nixai integration when available
      lua << EOF
      -- Load nixai integration if available
      local ok, nixai = pcall(require, 'nixai')
      if ok then
        nixai.setup({
          mcp_socket = os.getenv('HOME') .. '/.local/share/nixai/mcp.sock'
        })
        print('nixai integration loaded successfully')
      else
        print('nixai integration not available yet')
      end
      EOF
    ''';
  };

  # Enable nixai service with MCP server
  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      socketPath = \"\${config.home.homeDirectory}/.local/share/nixai/mcp.sock\";
      aiProvider = \"ollama\";
      aiModel = \"llama3\";
      host = \"localhost\";
      port = 8080;
    };
  };

  # Home Manager settings
  home.username = \"nixuser\";
  home.homeDirectory = \"/home/nixuser\";
  home.stateVersion = \"24.05\";

  # Enable home-manager
  programs.home-manager.enable = true;
}
EOF

        # Create the nixai module for Home Manager
        echo '📦 Creating nixai Home Manager module...'
        cp modules/home-manager.nix ~/.config/home-manager/nixai-module.nix
        
        # Install Home Manager
        echo '🏗️  Installing Home Manager...'
        nix-channel --add https://github.com/nix-community/home-manager/archive/release-24.05.tar.gz home-manager
        nix-channel --update
        
        # Install Home Manager and our configuration
        echo '⚙️  Applying Home Manager configuration...'
        nix-shell '<home-manager>' -A install || {
            echo '⚠️  Home Manager install failed, continuing with manual setup...'
            
            # Fallback: manual setup
            echo '🔧 Setting up nixai manually...'
            
            # Build nixai
            go mod tidy
            go build -o ./nixai ./cmd/nixai/main.go
            
            # Create Neovim configuration directory
            mkdir -p ~/.config/nvim/lua
            
            # Generate nixai Neovim module
            echo '📝 Generating nixai Neovim integration...'
            ./nixai neovim-setup || {
                echo '⚠️  Neovim setup command not available, creating manual setup...'
                
                # Create basic nixai Neovim integration
                cat > ~/.config/nvim/lua/nixai.lua << 'NVIM_EOF'
-- nixai Neovim integration module
local M = {}

-- Configuration
M.config = {
    mcp_socket = os.getenv('HOME') .. '/.local/share/nixai/mcp.sock',
    keymaps = true,
    auto_start_server = true
}

-- Start MCP server
function M.start_server()
    local socket_dir = vim.fn.fnamemodify(M.config.mcp_socket, ':h')
    vim.fn.mkdir(socket_dir, 'p')
    
    local cmd = string.format('nixai mcp-server start --socket-path=%s', M.config.mcp_socket)
    vim.fn.jobstart(cmd, {
        on_exit = function(_, code)
            if code == 0 then
                print('nixai MCP server started successfully')
            else
                print('Failed to start nixai MCP server')
            end
        end
    })
end

-- Query nixai
function M.query(question)
    if not question or question == '' then
        question = vim.fn.input('Ask nixai: ')
    end
    
    if question == '' then
        return
    end
    
    local cmd = string.format('nixai --ask \"%s\"', question)
    local output = vim.fn.system(cmd)
    
    -- Create a new buffer to show the response
    local buf = vim.api.nvim_create_buf(false, true)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, vim.split(output, '\n'))
    vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
    vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
    
    -- Open in a split window
    vim.cmd('split')
    vim.api.nvim_set_current_buf(buf)
end

-- Setup function
function M.setup(opts)
    M.config = vim.tbl_deep_extend('force', M.config, opts or {})
    
    if M.config.keymaps then
        -- Set up keymaps
        vim.keymap.set('n', '<leader>na', function() M.query() end, { desc = 'Ask nixai' })
        vim.keymap.set('v', '<leader>na', function()
            local start_pos = vim.fn.getpos(\"'<\")
            local end_pos = vim.fn.getpos(\"'>\")
            local lines = vim.fn.getline(start_pos[2], end_pos[2])
            local text = table.concat(lines, '\n')
            M.query('Explain this code: ' .. text)
        end, { desc = 'Ask nixai about selection' })
        
        vim.keymap.set('n', '<leader>ns', M.start_server, { desc = 'Start nixai MCP server' })
    end
    
    if M.config.auto_start_server then
        M.start_server()
    end
    
    print('nixai integration loaded')
end

return M
NVIM_EOF
                
                # Add to Neovim init.lua
                mkdir -p ~/.config/nvim
                echo \"require('nixai').setup()\" >> ~/.config/nvim/init.lua
            }
            
            # Start MCP server manually
            echo '🖥️  Starting nixai MCP server...'
            mkdir -p ~/.local/share/nixai
            ./nixai mcp-server start --socket-path=~/.local/share/nixai/mcp.sock &
            MCP_PID=\$!
            
            echo \"MCP server started with PID: \$MCP_PID\"
            
            # Test basic functionality
            echo '🧪 Testing nixai functionality...'
            ./nixai --ask \"How do I configure Neovim with Nix?\" || echo 'Direct query failed'
            
            # Test MCP server
            if [ -S ~/.local/share/nixai/mcp.sock ]; then
                echo '✅ MCP server socket created successfully'
            else
                echo '❌ MCP server socket not found'
            fi
        }
        
        echo ''
        echo '🎉 Setup complete! Available features:'
        echo '  • nixai binary: ./nixai --help'
        echo '  • Neovim with nixai integration: nvim'
        echo '  • MCP server for VS Code integration'
        echo '  • Home Manager modules for system integration'
        echo ''
        echo '📖 Try these commands:'
        echo '  ./nixai --ask \"How do I install packages in NixOS?\"'
        echo '  nvim (then use <leader>na to ask nixai questions)'
        echo '  ./nixai mcp-server start --socket-path=~/.local/share/nixai/mcp.sock'
        echo ''
        echo '🚀 Starting interactive shell...'
        
        # Start interactive shell
        exec bash
    "

log_success "Docker container with NixOS modules setup complete!"
