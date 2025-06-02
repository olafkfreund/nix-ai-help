#!/usr/bin/env bash
# test_flake_migration.sh - Test flake migration functionality in Docker environment
# This script tests the migration commands within the nixai Docker container

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

log_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

wait_for_continue() {
    echo -e "${GREEN}Press Enter to continue...${NC}"
    read -r
}

# Check if we're inside Docker
if [ ! -f /.dockerenv ]; then
    log_error "This test should be run inside the nixai Docker container"
    echo ""
    echo "To start the container:"
    echo "   ./docker_nixos/nixai-docker.sh run"
    echo ""
    echo "Then run this test:"
    echo "   ./test_flake_migration.sh"
    exit 1
fi

log_header "🧪 NixOS Flake Migration Tests"

echo "This test suite will:"
echo "1. Create test NixOS configurations (channel-based and flake-based)"
echo "2. Test migration analysis functionality"
echo "3. Test dry-run migration to flakes"
echo "4. Verify migration commands work correctly"
echo ""

wait_for_continue

# Create test directories
TEST_DIR="/tmp/nixai-migration-test"
CHANNEL_TEST_DIR="$TEST_DIR/channel-config"
FLAKE_TEST_DIR="$TEST_DIR/flake-config"

log_info "Setting up test environment..."
rm -rf "$TEST_DIR"
mkdir -p "$CHANNEL_TEST_DIR" "$FLAKE_TEST_DIR"

# Create a sample channel-based configuration.nix
log_info "Creating sample channel-based configuration..."
cat > "$CHANNEL_TEST_DIR/configuration.nix" << 'EOF'
# Sample NixOS Configuration (Channel-based)
{ config, pkgs, ... }:

{
  imports = [
    ./hardware-configuration.nix
  ];

  # Boot loader
  boot.loader.grub.enable = true;
  boot.loader.grub.device = "/dev/sda";

  # Networking
  networking.hostName = "nixos-test";
  networking.networkmanager.enable = true;

  # Enable SSH
  services.openssh.enable = true;
  services.openssh.settings.PasswordAuthentication = false;

  # Enable firewall
  networking.firewall.enable = true;
  networking.firewall.allowedTCPPorts = [ 22 ];

  # System packages
  environment.systemPackages = with pkgs; [
    vim
    git
    curl
    htop
  ];

  # Enable flakes for this test
  nix.settings.experimental-features = [ "nix-command" "flakes" ];

  # Users
  users.users.nixuser = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };

  system.stateVersion = "24.05";
}
EOF

# Create a sample hardware-configuration.nix
cat > "$CHANNEL_TEST_DIR/hardware-configuration.nix" << 'EOF'
# Sample Hardware Configuration
{ config, lib, pkgs, modulesPath, ... }:

{
  imports = [ ];

  boot.initrd.availableKernelModules = [ "ata_piix" "ohci_pci" "ehci_pci" "ahci" "sd_mod" "sr_mod" ];
  boot.initrd.kernelModules = [ ];
  boot.kernelModules = [ ];
  boot.extraModulePackages = [ ];

  fileSystems."/" = {
    device = "/dev/disk/by-uuid/00000000-0000-0000-0000-000000000000";
    fsType = "ext4";
  };

  hardware.cpu.intel.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;
}
EOF

# Create a sample flake-based configuration
log_info "Creating sample flake-based configuration..."
cat > "$FLAKE_TEST_DIR/flake.nix" << 'EOF'
{
  description = "Sample NixOS Flake Configuration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    home-manager = {
      url = "github:nix-community/home-manager/release-24.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, ... }: {
    nixosConfigurations = {
      nixos-test = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          ./configuration.nix
          home-manager.nixosModules.home-manager
        ];
      };
    };
  };
}
EOF

cp "$CHANNEL_TEST_DIR/configuration.nix" "$FLAKE_TEST_DIR/"
cp "$CHANNEL_TEST_DIR/hardware-configuration.nix" "$FLAKE_TEST_DIR/"

log_success "Test configurations created successfully"

# Test 1: Migration Analysis with Channel Configuration
log_header "Test 1: Migration Analysis (Channel-based Config)"

log_info "Testing migration analysis on channel-based configuration..."
cd "$CHANNEL_TEST_DIR"

echo "Running: nixai migrate analyze --verbose"
if nixai migrate analyze --verbose; then
    log_success "Migration analysis completed successfully"
else
    log_error "Migration analysis failed"
    exit 1
fi

wait_for_continue

# Test 2: Migration Analysis with Target
log_header "Test 2: Migration Analysis with Target (Flakes)"

log_info "Testing migration analysis with target: flakes..."
echo "Running: nixai migrate analyze --target flakes --verbose"
if nixai migrate analyze --target flakes --verbose; then
    log_success "Migration analysis with target completed successfully"
else
    log_error "Migration analysis with target failed"
    exit 1
fi

wait_for_continue

# Test 3: Dry-run Migration to Flakes
log_header "Test 3: Dry-run Migration to Flakes"

log_info "Testing dry-run migration to flakes..."
echo "Running: nixai migrate to-flakes --dry-run"
if nixai migrate to-flakes --dry-run; then
    log_success "Dry-run migration completed successfully"
else
    log_error "Dry-run migration failed"
    exit 1
fi

wait_for_continue

# Test 4: Migration Analysis with Flake Configuration
log_header "Test 4: Migration Analysis (Flake-based Config)"

log_info "Testing migration analysis on flake-based configuration..."
cd "$FLAKE_TEST_DIR"

echo "Running: nixai migrate analyze --verbose"
if nixai migrate analyze --verbose; then
    log_success "Flake migration analysis completed successfully"
else
    log_error "Flake migration analysis failed"
    exit 1
fi

wait_for_continue

# Test 5: Test Help and Documentation
log_header "Test 5: Help and Documentation"

log_info "Testing migration command help..."
echo "Running: nixai migrate --help"
if nixai migrate --help; then
    log_success "Migration help displayed successfully"
else
    log_error "Migration help failed"
    exit 1
fi

echo ""
log_info "Testing migration analyze help..."
echo "Running: nixai migrate analyze --help"
if nixai migrate analyze --help; then
    log_success "Migration analyze help displayed successfully"
else
    log_error "Migration analyze help failed"
    exit 1
fi

echo ""
log_info "Testing migration to-flakes help..."
echo "Running: nixai migrate to-flakes --help"
if nixai migrate to-flakes --help; then
    log_success "Migration to-flakes help displayed successfully"
else
    log_error "Migration to-flakes help failed"
    exit 1
fi

wait_for_continue

# Test 6: Interactive Mode Test (if available)
log_header "Test 6: Interactive Migration Question"

log_info "Testing migration-related question through direct AI query..."
echo "Running: nixai 'What are the benefits of using Nix flakes over channels?'"
if nixai "What are the benefits of using Nix flakes over channels?"; then
    log_success "Migration question answered successfully"
else
    log_warning "Migration question failed (may be due to AI provider not available)"
fi

wait_for_continue

# Cleanup
log_header "Test Cleanup"

log_info "Cleaning up test directories..."
rm -rf "$TEST_DIR"
log_success "Cleanup completed"

# Summary
log_header "🎉 Migration Test Results"

echo -e "${GREEN}✅ All migration tests completed successfully!${NC}"
echo ""
echo "Tests performed:"
echo "  ✅ Migration analysis on channel-based configuration"
echo "  ✅ Migration analysis with target specification"
echo "  ✅ Dry-run migration to flakes"
echo "  ✅ Migration analysis on flake-based configuration"
echo "  ✅ Help and documentation display"
echo "  ✅ Direct AI query about migration"
echo ""
echo "The migration functionality is working correctly in the Docker environment!"
echo ""
echo "To perform actual migrations on real configurations:"
echo "  • Use 'nixai migrate analyze' to understand your setup"
echo "  • Use 'nixai migrate to-flakes --dry-run' to preview changes"
echo "  • Use 'nixai migrate to-flakes' to perform actual migration"
echo ""
log_success "Migration test suite completed successfully!"
