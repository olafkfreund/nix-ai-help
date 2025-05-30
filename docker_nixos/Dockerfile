# Dockerfile for NixOS 25.05 Testing Environment
# This Dockerfile creates a fresh NixOS 25.05 container for testing nixai features, including MCP server, Neovim, and CLI integration.
#
# Usage:
#   docker build -t nixai-nixos-test .
#   docker run -it --rm nixai-nixos-test
#
# Note: Docker support for NixOS is experimental. This setup uses the official Nix installer and configures a minimal NixOS-like environment.

FROM ubuntu:24.04

LABEL maintainer="nixai contributors"
LABEL description="NixOS 25.05 test environment for nixai project"

# Install dependencies for Nix and Neovim
RUN apt-get update && \
    apt-get install -y curl sudo git locales neovim python3 python3-pip lsb-release && \
    locale-gen en_US.UTF-8

ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

# Install Nix (multi-user)
RUN useradd -m nixuser && \
    usermod -aG sudo nixuser && \
    echo "nixuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    mkdir -p /nix && chown nixuser /nix && \
    curl -L https://nixos.org/nix/install | sh -s -- --daemon && \
    chown -R nixuser:root /nix

# Enable experimental Nix features for all users
RUN echo 'experimental-features = nix-command flakes' >> /etc/nix/nix.conf

# Configure Git to trust the cloned repository directory
RUN git config --system --add safe.directory /home/nixuser/nixai

ENV USER=nixuser
ENV HOME=/home/nixuser
ENV NIX_PATH=nixpkgs=channel:nixos-25.05

# Create AI keys placeholder
RUN echo "# AI keys placeholder - add your keys here" > /etc/profile.d/ai_keys.sh && \
    chmod 644 /etc/profile.d/ai_keys.sh

# Set up Nix profile and channels for the user
USER nixuser
RUN . /etc/profile.d/nix.sh && \
    nix-channel --add https://nixos.org/channels/nixos-25.05 nixpkgs && \
    nix-channel --update

# Install NixOS tools, Neovim, Go, and other build dependencies
RUN . /etc/profile.d/nix.sh && \
    nix-env -iA nixpkgs.nixos-install-tools nixpkgs.neovim nixpkgs.git nixpkgs.curl nixpkgs.python3 nixpkgs.nodejs nixpkgs.alejandra nixpkgs.go

# Clone the nixai repository into /home/nixuser/nixai
RUN . /etc/profile.d/nix.sh && \
    cd /home/nixuser && \
    git clone https://github.com/olafkfreund/nix-ai-help.git nixai || \
    (echo "Warning: Could not clone remote repo, creating test structure..." && \
     mkdir -p nixai/cmd/nixai nixai/internal/config nixai/pkg/logger nixai/configs && \
     cd nixai && \
     echo 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("nixai test version")\n}' > cmd/nixai/main.go && \
     echo 'module nixai\n\ngo 1.21' > go.mod && \
     echo '# nixai Test Environment\n\nThis is a test environment for nixai development.' > README.md)

# Build nixai (if we have a proper Go project structure)
RUN . /etc/profile.d/nix.sh && \
    cd /home/nixuser/nixai && \
    if [ -f "go.mod" ]; then \
        go mod tidy && \
        go build -o ./nixai ./cmd/nixai/main.go && \
        sudo cp ./nixai /usr/local/bin/nixai && \
        sudo chmod +x /usr/local/bin/nixai; \
    else \
        echo "No go.mod found, skipping build"; \
    fi

# Set working directory to the cloned repository
WORKDIR /home/nixuser/nixai

# Ollama integration notes:
# To access your local Ollama instance from inside this container, use the special hostname 'host.docker.internal' (default Ollama port: 11434).
#
# Example: set OLLAMA_HOST in your environment or config to 'http://host.docker.internal:11434'.
#
# On Linux, you may need to add the following flag when running the container:
#   --add-host=host.docker.internal:host-gateway
#
# Example run command:
#   docker run -it --rm --add-host=host.docker.internal:host-gateway nixai-nixos-test
ENV OLLAMA_HOST=http://host.docker.internal:11434

# Entrypoint: start a shell with Nix and AI keys available
ENTRYPOINT ["/bin/bash", "-c", ". /etc/profile.d/nix.sh && . /etc/profile.d/ai_keys.sh && exec bash"]
