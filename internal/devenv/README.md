# Devenv Integration in nixai

This document explains the development environment (devenv) integration in nixai, how to use it, and how to add support for new languages or frameworks.

---

## 🚀 What is the Devenv Feature?

nixai's `devenv` feature lets you quickly scaffold reproducible development environments for popular languages (Python, Rust, Node.js, Go) using [devenv.sh](https://devenv.sh/) and Nix. It provides:

- **Language templates** for Python, Rust, Node.js, Go (with more coming)
- **Framework and tool options** (e.g., Flask, FastAPI, Gin, gRPC, TypeScript, etc.)
- **Database/service integration** (Postgres, Redis, MySQL, MongoDB)
- **AI-powered template suggestion**
- **Extensible template system** for adding new languages

---

## 🧑‍💻 Usage Examples

### List Available Templates

```sh
nixai devenv list
```

### Create a New Project

```sh
nixai devenv create python myproject --framework fastapi --with-poetry --services postgres,redis
nixai devenv create golang my-go-app --framework gin --with-grpc
nixai devenv create nodejs my-node-app --with-typescript --services mongodb
nixai devenv create rust my-rust-app --with-wasm
```

### Get AI-Powered Suggestions

```sh
nixai devenv suggest "web app with database and REST API"
```

---

## 🛠️ How It Works

- All templates are implemented in Go in `builtin_templates.go` (see this file for examples).
- Each template implements the `Template` interface (see `plugin.go`).
- The CLI parses user options and passes them to the template's `Generate` method.
- The generated `devenv.nix` file is written to your project directory.

---

## 🧩 Adding Support for a New Language or Framework

1. **Edit `builtin_templates.go`**
   - Copy an existing template struct (e.g., `PythonTemplate`)
   - Implement the `Template` interface:
     - `Name() string`
     - `Description() string`
     - `RequiredInputs() []InputField`
     - `SupportedServices() []string`
     - `Validate(config TemplateConfig) error`
     - `Generate(config TemplateConfig) (*DevenvConfig, error)`
2. **Register your template**
   - Add it to the list in `registerBuiltinTemplates()` in `service.go`
3. **Test**
   - Add or update tests in `service_test.go`
4. **Document**
   - Update this README and the main project docs with your new template

---

## 📝 Example: Minimal Template Implementation

```go
// ExampleTemplate implements the Template interface
 type ExampleTemplate struct{}

 func (e *ExampleTemplate) Name() string { return "example" }
 func (e *ExampleTemplate) Description() string { return "Example language environment" }
 func (e *ExampleTemplate) RequiredInputs() []devenv.InputField { return nil }
 func (e *ExampleTemplate) SupportedServices() []string { return nil }
 func (e *ExampleTemplate) Validate(config devenv.TemplateConfig) error { return nil }
 func (e *ExampleTemplate) Generate(config devenv.TemplateConfig) (*devenv.DevenvConfig, error) {
     // ... generate config ...
     return &devenv.DevenvConfig{/* ... */}, nil
 }
```

---

## 🧪 Testing

- Run all tests: `go test ./internal/devenv/...`
- Try creating projects with various options and check the generated `devenv.nix`

---

## 📚 References

- [devenv.sh documentation](https://devenv.sh/)
- [nix.dev](https://nix.dev/)
- [NixOS Manual](https://nixos.org/manual/nixpkgs/stable/)

---

For questions or contributions, see the main project README.
