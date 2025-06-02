# Dockerfile - NixOS Docker container for nixai testing
# This uses an official NixOS Docker image with proper module system support

FROM nixos/nix:latest

LABEL maintainer="nixai contributors"
LABEL description="NixOS environment for testing nixai functionality"

# Create working directory and set up environment
# We'll run as root for simplicity in this test environment
WORKDIR /root

# Enable experimental Nix features
RUN mkdir -p /etc/nix && \
    echo 'experimental-features = nix-command flakes' >> /etc/nix/nix.conf && \
    echo 'trusted-users = root' >> /etc/nix/nix.conf

# Set up Nix channels  
RUN nix-channel --add https://nixos.org/channels/nixos-unstable nixpkgs && \
    nix-channel --update

# Install essential packages directly via nix-env (no Home Manager conflicts)
RUN nix-env -iA nixpkgs.git \
    nixpkgs.curl \
    nixpkgs.go \
    nixpkgs.alejandra \
    nixpkgs.neovim \
    nixpkgs.just \
    nixpkgs.tree \
    nixpkgs.jq \
    nixpkgs.sed \
    nixpkgs.inetutils \
    nixpkgs.netcat \
    nixpkgs.host \
    nixpkgs.iproute2

# Copy nixai source code
COPY . nixai

# Copy AI keys file separately for better layer caching
COPY ./docker_nixos/.ai_keys /root/.ai_keys
COPY ./docker_nixos/test_host_connectivity.sh /root/test_host_connectivity.sh
RUN chmod +x /root/test_host_connectivity.sh

# Set working directory
WORKDIR /root/nixai

# Build nixai
RUN nix develop --command bash -c "go mod tidy && go build -o ./nixai ./cmd/nixai/main.go"

# Set up Ollama environment for Docker
ENV OLLAMA_HOST=http://host.docker.internal:11434

# Create entrypoint script
RUN echo '#!/usr/bin/env bash' > /root/entrypoint.sh && \
    echo 'set -e' >> /root/entrypoint.sh && \
    echo '' >> /root/entrypoint.sh && \
    echo '# Source AI API keys' >> /root/entrypoint.sh && \
    echo 'if [ -f "/root/.ai_keys" ]; then' >> /root/entrypoint.sh && \
    echo '  . /root/.ai_keys' >> /root/entrypoint.sh && \
    echo '  echo "🔑 AI API keys loaded"' >> /root/entrypoint.sh && \
    echo 'fi' >> /root/entrypoint.sh && \
    echo '' >> /root/entrypoint.sh && \
    echo 'echo "🚀 NixOS Docker Environment for nixai Testing"' >> /root/entrypoint.sh && \
    echo 'echo "=============================================="' >> /root/entrypoint.sh && \
    echo 'echo ""' >> /root/entrypoint.sh && \
    echo 'echo "📦 Available components:"' >> /root/entrypoint.sh && \
    echo 'echo "  • nixai binary: ~/nixai/nixai"' >> /root/entrypoint.sh && \
    echo 'echo "  • Neovim editor"' >> /root/entrypoint.sh && \
    echo 'echo "  • NixOS tools and packages"' >> /root/entrypoint.sh && \
    echo 'echo "  • MCP server support"' >> /root/entrypoint.sh && \
    echo 'echo ""' >> /root/entrypoint.sh && \
    echo 'echo "🔧 Quick start:"' >> /root/entrypoint.sh && \
    echo 'echo "  cd ~/nixai"' >> /root/entrypoint.sh && \
    echo 'echo "  ./nixai --ask \"How do I configure NixOS?\""' >> /root/entrypoint.sh && \
    echo 'echo "  ./nixai explain-option services.ssh.enable"' >> /root/entrypoint.sh && \
    echo 'echo "  nvim  # Opens Neovim"' >> /root/entrypoint.sh && \
    echo 'echo ""' >> /root/entrypoint.sh && \
    echo 'echo "📝 Test commands:"' >> /root/entrypoint.sh && \
    echo 'echo "  ./nixai --help"' >> /root/entrypoint.sh && \
    echo 'echo "  ./nixai mcp-server start --help"' >> /root/entrypoint.sh && \
    echo 'echo ""' >> /root/entrypoint.sh && \
    echo '' >> /root/entrypoint.sh && \
    echo '# Check Ollama connectivity' >> /root/entrypoint.sh && \
    echo 'echo "🔄 Checking Ollama connectivity..."' >> /root/entrypoint.sh && \
    echo 'if curl -s --connect-timeout 5 "${OLLAMA_HOST}/api/version" &>/dev/null; then' >> /root/entrypoint.sh && \
    echo '  echo "✅ Ollama server is reachable at ${OLLAMA_HOST}"' >> /root/entrypoint.sh && \
    echo '  OLLAMA_VERSION=$(curl -s "${OLLAMA_HOST}/api/version" | grep -o '"'"'"version":"[^"]*"'"'"' | sed '"'"'s/"version":"//'"'"' | sed '"'"'s/"//g'"'"' || echo "unknown")' >> /root/entrypoint.sh && \
    echo '  echo "   Ollama version: ${OLLAMA_VERSION}"' >> /root/entrypoint.sh && \
    echo 'else' >> /root/entrypoint.sh && \
    echo '  echo "⚠️  Warning: Cannot connect to Ollama server at ${OLLAMA_HOST}"' >> /root/entrypoint.sh && \
    echo '  echo "   This might affect some nixai functionality."' >> /root/entrypoint.sh && \
    echo '  echo "   Run ./docker_nixos/test_host_connectivity.sh to diagnose connection issues."' >> /root/entrypoint.sh && \
    echo 'fi' >> /root/entrypoint.sh && \
    echo '' >> /root/entrypoint.sh && \
    echo '# Start MCP server in background' >> /root/entrypoint.sh && \
    echo 'cd /root/nixai' >> /root/entrypoint.sh && \
    echo 'echo "🖥️  Starting nixai MCP server..."' >> /root/entrypoint.sh && \
    echo './nixai mcp-server start --socket-path=/root/.local/share/nixai/mcp.sock --background || echo "MCP server start failed (this is OK for testing)"' >> /root/entrypoint.sh && \
    echo '' >> /root/entrypoint.sh && \
    echo 'echo "🎉 Ready! Starting interactive shell..."' >> /root/entrypoint.sh && \
    echo 'exec "$@"' >> /root/entrypoint.sh

RUN chmod +x /root/entrypoint.sh

# Set environment
ENV HOME=/root
ENV USER=root

# Use the correct path to bash in the NixOS container
ENTRYPOINT ["/root/.nix-profile/bin/bash", "/root/entrypoint.sh"]
CMD []
