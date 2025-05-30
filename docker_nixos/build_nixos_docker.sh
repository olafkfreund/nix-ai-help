#!/usr/bin/env bash
# build_nixos_docker.sh - Build and run nixai with full NixOS Docker support

set -euo pipefail

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

log_section() {
    echo -e "${CYAN}[SECTION]${NC} $1"
    echo "────────────────────────────────────────────────"
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker first."
    exit 1
fi

log_section "Building NixOS Docker Image with nixai Modules"

# Build the NixOS Docker image
log_info "Building nixai NixOS Docker image..."
docker build -t nixai-nixos-full -f docker_nixos/Dockerfile .. || {
    log_error "Docker build failed!"
    exit 1
}

# Check if container already exists and stop it
CONTAINER_NAME="nixai-nixos-test"
if docker ps -a --format '{{.Names}}' | grep -q "$CONTAINER_NAME"; then
    log_warning "Stopping existing container..."
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

log_section "Starting NixOS Container with Full Module Support"

# Run the container with proper networking for Ollama
log_info "Starting nixai NixOS container..."
docker run -it --name "$CONTAINER_NAME" \
    --add-host=host.docker.internal:host-gateway \
    -e OLLAMA_HOST=http://host.docker.internal:11434 \
    nixai-nixos-full

log_success "NixOS Docker container setup complete!"
