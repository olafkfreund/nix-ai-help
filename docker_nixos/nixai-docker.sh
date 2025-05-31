#!/usr/bin/env bash
# nixai-docker.sh - Build and run nixai with NixOS Docker support
# This script handles all Docker operations for nixai testing

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

# Default values
IMAGE_NAME="nixai-nixos-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONTAINER_NAME="nixai-docker-test"
SKIP_BUILD=false
COMMAND=""

# Detect platform
PLATFORM=$(uname -s)
log_info "Detected platform: $PLATFORM"

# Default Ollama host configuration
OLLAMA_HOST_URL="http://host.docker.internal:11434"
EXTRA_DOCKER_ARGS=()

# Setup for host.docker.internal mapping
case "$PLATFORM" in
    Linux)
        log_info "Linux platform detected - adding host gateway mapping"
        # Linux needs special handling for host.docker.internal
        EXTRA_DOCKER_ARGS+=("--add-host=host.docker.internal:host-gateway")
        ;;
    Darwin)
        log_info "macOS platform detected - host.docker.internal should work natively"
        # macOS Docker Desktop supports host.docker.internal natively
        ;;
    *)
        log_warning "Unknown platform - using default Docker configuration"
        EXTRA_DOCKER_ARGS+=("--add-host=host.docker.internal:host-gateway")
        ;;
esac

# Prepare necessary files

# Create .ai_keys if it doesn't exist
if [ ! -f "$SCRIPT_DIR/.ai_keys" ]; then
    log_info "Creating sample .ai_keys file"
    cat > "$SCRIPT_DIR/.ai_keys" << EOT
# AI Provider API Keys
# Replace with your actual keys if needed
export OPENAI_API_KEY=""
export GEMINI_API_KEY=""
export OLLAMA_HOST="http://host.docker.internal:11434"
EOT
    log_success "Created sample .ai_keys file"
else
    log_info ".ai_keys file already exists"
fi

# Create test_host_connectivity.sh if it doesn't exist
if [ ! -f "$SCRIPT_DIR/test_host_connectivity.sh" ] || [ ! -s "$SCRIPT_DIR/test_host_connectivity.sh" ]; then
    log_info "Creating test_host_connectivity.sh"
    cat > "$SCRIPT_DIR/test_host_connectivity.sh" << 'EOT'
#!/usr/bin/env bash
# test_host_connectivity.sh - Tests Docker host connectivity for Ollama
# This script tests if the container can connect to the Ollama server on the host

set -e

# Print header
echo "🔍 Testing Ollama host connectivity..."
echo "=================================="

# Check if host.docker.internal resolves
echo "Testing host.docker.internal DNS resolution..."
if host host.docker.internal &>/dev/null; then
  echo "✅ host.docker.internal resolves successfully"
  host host.docker.internal
else
  echo "❌ host.docker.internal does not resolve"
  echo "Fallback: Adding explicit DNS entry for Linux hosts..."
  DOCKER_HOST_IP=$(ip route | grep default | awk '{print $3}')
  echo "Docker host IP detected as: $DOCKER_HOST_IP"
  echo "$DOCKER_HOST_IP host.docker.internal" >> /etc/hosts
  echo "Added host.docker.internal to /etc/hosts pointing to $DOCKER_HOST_IP"
  host host.docker.internal || echo "Still cannot resolve host.docker.internal"
fi

# Check if Ollama port is reachable
echo -e "\nTesting Ollama API endpoint..."
if curl -s --connect-timeout 5 http://host.docker.internal:11434/api/version &>/dev/null; then
  echo "✅ Ollama API is reachable at http://host.docker.internal:11434"
  echo "Version: $(curl -s http://host.docker.internal:11434/api/version)"
else
  echo "❌ Cannot reach Ollama API at http://host.docker.internal:11434"
  echo "Checking if port 11434 is open on host.docker.internal..."
  nc -zv host.docker.internal 11434 || echo "Port 11434 is not open or reachable"
fi

# Check environment variable configuration
echo -e "\nChecking Ollama environment configuration..."
echo "OLLAMA_HOST=$OLLAMA_HOST"

# Test connection using the configured OLLAMA_HOST
if [[ -n "$OLLAMA_HOST" ]]; then
  echo "Testing connection to configured OLLAMA_HOST: $OLLAMA_HOST"
  if curl -s --connect-timeout 5 "$OLLAMA_HOST/api/version" &>/dev/null; then
    echo "✅ Ollama API is reachable at $OLLAMA_HOST"
    echo "Version: $(curl -s "$OLLAMA_HOST/api/version")"
  else
    echo "❌ Cannot reach Ollama API at $OLLAMA_HOST"
  fi
fi

# Check for other potential issues
echo -e "\nChecking for potential networking issues..."
echo "Container's outbound IP: $(curl -s https://ifconfig.me || echo "N/A")"
echo "Network interfaces:"
ip addr | grep -E 'inet|eth|docker'

echo -e "\n🔍 Connectivity test complete"
EOT
    chmod +x "$SCRIPT_DIR/test_host_connectivity.sh"
    log_success "Created test_host_connectivity.sh"
else
    log_info "test_host_connectivity.sh already exists"
fi

# Process command line arguments
show_usage() {
    echo "Usage: $0 [options] [command]"
    echo ""
    echo "Options:"
    echo "  --no-build         Skip building the Docker image"
    echo "  --help             Show this help message"
    echo "  --demo             Run container and start the interactive demo"
    echo ""
    echo "Commands:"
    echo "  build              Only build the Docker image"
    echo "  run                Run the container (default if no command specified)"
    echo "  test               Run comprehensive test suite inside container"
    echo "  shell              Start an interactive shell (default)"
    echo "  demo               Run the interactive nixai features demo"
    echo ""
    echo "Examples:"
    echo "  $0                 Build and run with interactive shell"
    echo "  $0 build           Only build the Docker image"
    echo "  $0 --no-build run  Run without rebuilding the image"
    echo "  $0 test            Run tests inside container"
    echo "  $0 demo            Run the interactive demo"
    echo "  $0 --demo          Same as 'demo' command"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-build)
            SKIP_BUILD=true
            shift
            ;;
        --demo)
            COMMAND="demo"
            shift
            ;;
        --help)
            show_usage
            ;;
        build|run|test|shell|demo)
            COMMAND=$1
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            ;;
    esac
done

# If no command specified, default to "run"
if [[ -z "$COMMAND" ]]; then
    COMMAND="run"
fi

# Check if Ollama is running before starting, if we're going to run a container
if [[ "$COMMAND" != "build" ]]; then
    log_info "Checking if Ollama is running on host..."
    if curl -s --connect-timeout 2 "http://localhost:11434/api/version" &>/dev/null; then
        OLLAMA_VERSION=$(curl -s "http://localhost:11434/api/version" | grep -o '"version":"[^"]*"' || echo "unknown")
        log_success "Ollama is running - Version: $OLLAMA_VERSION"
    else
        log_warning "Ollama does not appear to be running on host."
        log_info "It's recommended to start Ollama with: ollama serve"
        log_info "Continuing anyway, but some features may not work..."
    fi
fi

# Build the Docker image (unless --no-build specified)
if [[ "$SKIP_BUILD" = false ]] || [[ "$COMMAND" = "build" ]]; then
    log_info "Building Docker image: $IMAGE_NAME"
    docker build -t "$IMAGE_NAME" -f "$SCRIPT_DIR/Dockerfile" "$PROJECT_ROOT"
    log_success "Docker image built: $IMAGE_NAME"
fi

# Handle different commands
case "$COMMAND" in
    build)
        log_success "Docker image build complete!"
        ;;
    test)
        log_info "Running test suite in container..."
        docker run -it --rm \
            --name "$CONTAINER_NAME" \
            "${EXTRA_DOCKER_ARGS[@]}" \
            -e OLLAMA_HOST="$OLLAMA_HOST_URL" \
            "$IMAGE_NAME" bash -c "cd /root/nixai && ./docker_nixos/test_docker_nixai.sh"
        ;;
    demo)
        log_info "Running nixai features demo in container..."
        # Remove existing container if it exists
        if docker container inspect "$CONTAINER_NAME" &>/dev/null 2>&1; then
            log_info "Removing existing container: $CONTAINER_NAME"
            docker rm -f "$CONTAINER_NAME" >/dev/null
        fi

        log_info "Starting demo container..."
        docker run -it --rm \
            --name "$CONTAINER_NAME" \
            "${EXTRA_DOCKER_ARGS[@]}" \
            -e OLLAMA_HOST="$OLLAMA_HOST_URL" \
            "$IMAGE_NAME" bash -c "cd /root/nixai && ./docker_nixos/demo_nixai_features.sh"
        ;;
    run|shell)
        log_info "Starting container with interactive shell..."
        # Remove existing container if it exists
        if docker container inspect "$CONTAINER_NAME" &>/dev/null 2>&1; then
            log_info "Removing existing container: $CONTAINER_NAME"
            docker rm -f "$CONTAINER_NAME" >/dev/null
        fi

        log_info "Running container with Ollama connectivity..."
        docker run -it --rm \
            --name "$CONTAINER_NAME" \
            "${EXTRA_DOCKER_ARGS[@]}" \
            -e OLLAMA_HOST="$OLLAMA_HOST_URL" \
            "$IMAGE_NAME"
        ;;
esac

log_success "Docker operation complete!"
