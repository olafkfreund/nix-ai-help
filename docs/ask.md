# nixai ask

Ask any NixOS-related question and get an AI-powered answer.

---

## Command Help Output

```sh
./nixai ask --help
Ask a direct question about NixOS configuration and get an AI-powered answer with comprehensive multi-source validation.

This command queries multiple information sources:
- Official NixOS documentation via MCP server
- Verified package search results
- Real-world GitHub configuration examples
- Response validation for common syntax errors

Use --quiet to suppress validation output and show only the AI response.

Usage:
  nixai ask [question] [flags]

Aliases:
  ask, a

Flags:
  -h, --help    help for ask
  -q, --quiet   Suppress validation output and show only the AI response

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai ask "How do I configure nginx?"
  nixai ask "What is the difference between services.openssh.enable and programs.ssh.enable?"
  nixai ask "How do I set up a development environment with Python?"
  nixai ask "Generate a secure NixOS configuration" # Uses Claude for complex reasoning
  nixai ask "Quick help with flakes setup" # Uses Groq for fast response
  nixai ask "How do I enable SSH?" --quiet
```

---

## The `--quiet` Flag

The `--quiet` (or `-q`) flag provides a streamlined experience by suppressing all validation output and progress indicators, showing only the final AI response. This is useful for:

- **Scripting**: When you need just the AI response for automation
- **Clean Output**: When you want minimal, distraction-free answers
- **Performance**: Slightly faster execution by skipping validation displays

### Normal vs Quiet Mode

**Normal mode** (default):
- Shows multi-source information gathering progress
- Displays validation from official documentation
- Shows package search results
- Includes real-world configuration examples  
- Provides quality indicators and tips
- Full verbose output with context

**Quiet mode** (`--quiet` or `-q`):
- Silent information gathering (no progress output)
- Only displays the final AI response
- All the same AI intelligence and sources, just minimal output
- Perfect for piping or scripting usage

Both modes use the same comprehensive AI analysis - quiet mode just changes the output presentation.

---

## Real Life Examples

- **Ask about enabling a service:**

  ```sh
  nixai "How do I enable SSH in NixOS?"
  ```

- **Ask about troubleshooting:**

  ```sh
  nixai --ask "Why is my system not booting after an update?"
  ```

- **Ask with quiet mode (only AI response, no validation output):**

  ```sh
  nixai ask "How do I configure nginx?" --quiet
  nixai ask "What's the difference between flakes and channels?" -q
  ```

- **Ask with specific AI provider:**

  ```sh
  nixai ask "How do I set up Home Manager?" --provider gemini
  nixai ask "Debug my NixOS build failure" --provider openai --quiet
  ```
