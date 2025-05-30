#!/usr/bin/env bash
# build_and_run_docker.sh
# Build the nixai-nixos-test Docker image with isolated nixai repository and start a container.
# Usage: ./build_and_run_docker.sh [additional docker run args]

set -euo pipefail

IMAGE_NAME="nixai-nixos-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build the Docker image from the docker_nixos directory
echo "[+] Building Docker image: $IMAGE_NAME"
echo "    This will create an isolated NixOS 25.05 environment with nixai cloned inside"
docker build -t "$IMAGE_NAME" "$SCRIPT_DIR"

echo "[+] Starting Docker container with nixai environment..."
echo "    - Isolated nixai repository cloned in /home/nixuser/nixai"
echo "    - Adds host.docker.internal for Ollama access"
echo "    - NixOS 25.05 environment with all development tools"
echo "    - Interactive shell with Nix, Neovim, and development environment"
echo ""
echo "🧪 To run the comprehensive test suite inside the container:"
echo "    cd /home/nixuser/nixai && ./docker_nixos/test_docker_nixai.sh"
echo ""
echo "🚀 Quick start commands inside the container:"
echo "    nixai --help"
echo "    nixai 'How do I configure git in NixOS?'"
echo "    nix develop .#docker"
echo "    just build && just test"
echo ""

# Handle special commands
case "${1:-}" in
    "test")
        echo "🧪 Running comprehensive test suite..."
        docker run -it --rm \
          --add-host=host.docker.internal:host-gateway \
          "$IMAGE_NAME" bash -c "cd /home/nixuser/nixai && ./docker_nixos/test_docker_nixai.sh"
        ;;
    "demo")
        echo "🚀 Running feature demonstration..."
        docker run -it --rm \
          --add-host=host.docker.internal:host-gateway \
          "$IMAGE_NAME" bash -c "cd /home/nixuser/nixai && ./docker_nixos/demo_nixai_features.sh"
        ;;
    "shell"|"")
        echo "🐚 Starting interactive shell..."
        docker run -it --rm \
          --add-host=host.docker.internal:host-gateway \
          "$IMAGE_NAME"
        ;;
    *)
        echo "🔧 Running custom command: $*"
        docker run -it --rm \
          --add-host=host.docker.internal:host-gateway \
          "$IMAGE_NAME" "$@"
        ;;
esac
