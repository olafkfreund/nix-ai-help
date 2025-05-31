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
