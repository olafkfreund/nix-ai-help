# NixAI Flake Quick Reference

## Features

- Multi-system support (x86_64-linux, aarch64-linux, darwin)
- Reproducible Go build for `nixai` CLI
- Dev shells for local and Docker development
- NixOS and Home Manager modules
- Built-in code formatter (alejandra)
- Lint check for Go code (golangci-lint)

## Usage Examples

### Build the CLI

```zsh
nix build .#nixai
```

### Run the CLI

```zsh
nix run .#nixai -- --help
```

### Enter Dev Shell

```zsh
nix develop
```

### Run Lint Check

```zsh
nix flake check
```

### Format Nix Code

```zsh
nix fmt
```

### Use NixOS Module

Add to your configuration:

```nix
imports = [
  (fetchGit { url = "https://github.com/olafkfreund/nix-ai-help"; }) + "/modules/nixos.nix"
];
```

### Use Home Manager Module

Add to your configuration:

```nix
imports = [
  (fetchGit { url = "https://github.com/olafkfreund/nix-ai-help"; }) + "/modules/home-manager.nix"
];
```

---

See this file and `docs/FLAKE_INTEGRATION_GUIDE.md` for more details.
