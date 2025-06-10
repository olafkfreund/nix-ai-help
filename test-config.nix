{ config, pkgs, ... }:

# System-level service configurations
services = {
  # Enable the systemd-based init system
  systemd.enable = true;

  # Start the display manager (e.g., gdm3)
  display-manager.enable = true;
  display-manager.displayManager = "gdm3";

  # Start the window manager (e.g., i3)
  window-manager.enable = true;
  window-manager.windowManager = "i3";
};

# Hardware enablement
hardware = {
  # Enable the graphics driver for your GPU (e.g., intel)
  graphics.enable = true;

  # Enable USB support
  usb.enable = true;

  # Enable Bluetooth support
  bluetooth.enable = true;
};

# Security and networking settings
security = {
  # Use AppArmor to enforce security policies
  apparmor.enable = true;

  # Disable IPv6 by default (you can re-enable it if needed)
  networking.ipv6.enable = false;

  # Set the default firewall configuration
  firewalld.enable = true;
};

# Package installations
packages = {
  # Install the base package set (including the X11 window system and KDE)
  basePackages = [ pkgs.kde Plasma ];

  # Install additional packages (e.g., chromium, libreoffice, git)
  extraPackages = [
    pkgs.chromium
    pkgs.libreoffice
    pkgs.git
  ];
};

# User and group configurations
users = {
  # Create a user account (replace "<username>" with your desired username)
  users = {
    <username> = {
      isNormalUser = true;
      extraGroups = [ "wheel" ]; // Add the user to the wheel group for sudo access
    };
  };

  # Configure the default shell (e.g., bash, zsh, fish)
  defaultShell = pkgs.bash;
};

# Advanced configuration options

# Performance optimizations
nix = {
  # Enable parallel builds for faster compilation times
  parallelBuilds = true;

  # Set a reasonable maximum number of jobs for parallel builds
  maxJobs = 4;
};

# Modular configuration structure
modules = [
  { config, pkgs, ... }: {
    # Add additional modules here (e.g., home-manager)
    imports = [ <home-manager-module> ];
  };
];

# Comprehensive documentation and comments

# Troubleshooting notes

# Error handling and fallbacks

# Alternative configuration options as comments
#
# (Replace "<username>" with your desired username)
# users.users.<username>.extraGroups = [ "wheel" ]; // Add the user to the wheel group for sudo access
2. Hardware enablement:
	* Disables IPv6 by default (you can re-enable it if needed).