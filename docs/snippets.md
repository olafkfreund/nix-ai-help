# nixai snippets

Access and manage configuration snippets and templates.

---

## Command Help Output

```sh
./nixai snippets --help
Access and manage configuration snippets and templates.

Usage:
  nixai snippets [command]

Available Commands:
  list      List all snippets
  show      Show a snippet for a given topic
  add       Add a new snippet
  remove    Remove a snippet

Flags:
  -h, --help   help for snippets

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai snippets list
  nixai snippets show nginx
```

---

## Real Life Examples

- **List all available snippets:**
  ```sh
  nixai snippets list
  # Shows all configuration snippets
  ```
- **Show a snippet for nginx:**
  ```sh
  nixai snippets show nginx
  # Displays a ready-to-use nginx config snippet
  ```
- **Add a new snippet:**
  ```sh
  nixai snippets add my-snippet
  # Adds a new snippet to the collection
  ```
