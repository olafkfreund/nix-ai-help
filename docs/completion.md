# nixai completion

Generate shell completion scripts for your shell.

---

## Command Help Output

```sh
./nixai completion --help
Generate shell completion scripts for your shell.

Usage:
  nixai completion [bash|zsh|fish]

Flags:
  -h, --help   help for completion

Examples:
  nixai completion zsh
  nixai completion bash
  nixai completion fish
```

---

## Real Life Examples

- **Enable zsh completion for nixai:**
  ```sh
  nixai completion zsh > ~/.oh-my-zsh/completions/_nixai
  source ~/.oh-my-zsh/completions/_nixai
  ```
- **Enable bash completion for nixai:**
  ```sh
  nixai completion bash > /etc/bash_completion.d/nixai
  source /etc/bash_completion.d/nixai
  ```
