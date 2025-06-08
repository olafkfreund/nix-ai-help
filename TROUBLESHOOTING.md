# Troubleshooting nixai Build Issues

## "go.mod file not found" Error

If you're getting an error like:
```
go: go.mod file not found in current directory or any parent directory
```

This typically happens when the source archive doesn't contain the `go.mod` file. Here are the solutions:

### Solution 1: Use the Flake (Recommended)

Instead of importing the NixOS module directly, use the flake:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai = {
      url = "github:olafkfreund/nix-ai-help";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, nixai, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        nixai.nixosModules.default
        {
          services.nixai.enable = true;
          services.nixai.mcp.enable = true;
        }
      ];
    };
  };
}
```

### Solution 2: Clone the Repository

If you want to use the module locally:

```bash
# Clone the repository
git clone https://github.com/olafkfreund/nix-ai-help.git /etc/nixos/nixai

# Import the module in your configuration.nix
imports = [
  /etc/nixos/nixai/modules/nixos.nix
];

services.nixai.enable = true;
```

### Solution 3: Direct Package Installation

You can also install nixai directly without the module:

```nix
environment.systemPackages = with pkgs; [
  (pkgs.callPackage (pkgs.fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "main"; # or specific commit
    sha256 = "sha256-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX";
  } + "/package.nix") {})
];
```

### Solution 4: Verify Source Contents

If you're building from source, make sure the source contains `go.mod`:

```bash
# Check if go.mod exists in your source
ls -la go.mod go.sum

# If missing, you might have an incomplete source archive
```

### Getting Help

If none of these solutions work:

1. Check that you're using the latest version
2. Verify your Nix version: `nix --version`
3. Try rebuilding with: `nixos-rebuild switch --show-trace`
4. Open an issue at: https://github.com/olafkfreund/nix-ai-help/issues

### For Developers

If you're developing nixai and get this error:

```bash
# Make sure go.mod exists and is committed
git add go.mod go.sum
git commit -m "Add Go module files"

# Clean build
rm -f result
nix build
```

The `vendorHash` in `package.nix` is: `sha256-pGyNwzTkHuOzEDOjmkzx0sfb1jHsqb/1FcojsCGR6CY=`
