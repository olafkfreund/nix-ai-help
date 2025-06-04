# nixai package-repo

Analyze a Git repository and generate a Nix derivation for packaging.

---

## Command Help Output

```sh
./nixai package-repo --help
Analyze a Git repository and generate a Nix derivation for packaging.

Usage:
  nixai package-repo <repo-url>

Flags:
  -h, --help   help for package-repo
  --output FILE   Write the generated derivation to FILE

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai package-repo https://github.com/user/project
  nixai package-repo https://github.com/user/project --output default.nix
```

---

## Usage

```sh
nixai package-repo <repo-url>
```

---

## Real Life Examples

- **Generate a Nix derivation for a GitHub project:**
  ```sh
  nixai package-repo https://github.com/user/project
  # Analyzes the repo and creates a default.nix
  ```
- **Write the derivation to a specific file:**
  ```sh
  nixai package-repo https://github.com/user/project --output my-derivation.nix
  # Saves the generated derivation to my-derivation.nix
  ```
