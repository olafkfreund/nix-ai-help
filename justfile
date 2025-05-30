# Justfile for nixai project
set shell := ["bash", "-c"]

# Build the application
build:
	@echo "Building nixai..."
	go build -o ./nixai ./cmd/nixai/main.go

# Build for production with optimizations
build-prod:
	@echo "Building nixai for production..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o ./nixai ./cmd/nixai/main.go

# Build for multiple architectures
build-all:
	@echo "Building nixai for multiple architectures..."
	GOOS=linux GOARCH=amd64 go build -o ./dist/nixai-linux-amd64 ./cmd/nixai/main.go
	GOOS=linux GOARCH=arm64 go build -o ./dist/nixai-linux-arm64 ./cmd/nixai/main.go
	GOOS=darwin GOARCH=amd64 go build -o ./dist/nixai-darwin-amd64 ./cmd/nixai/main.go
	GOOS=darwin GOARCH=arm64 go build -o ./dist/nixai-darwin-arm64 ./cmd/nixai/main.go

# Run the application
run: build
	@echo "Running nixai..."
	./nixai

# Run with specific arguments
run-args ARGS: build
	@echo "Running nixai with arguments: {{ARGS}}"
	./nixai {{ARGS}}

# Run in interactive mode
run-interactive: build
	@echo "Running nixai in interactive mode..."
	./nixai --interactive

# Run MCP server
run-mcp: build
	@echo "Starting MCP server..."
	./nixai mcp-server start

# Run with debug logging
run-debug: build
	@echo "Running nixai with debug logging..."
	./nixai --log-level debug

# Test the application
test:
	@echo "Running tests..."
	go test ./...

# Run all tests (including Go tests, MCP tests, VS Code integration tests, provider tests)
test-all:
	@echo "Running all tests..."
	./tests/run_all.sh

# Run only MCP tests
test-mcp:
	@echo "Running MCP tests..."
	./tests/run_mcp.sh

# Run only VS Code integration tests
test-vscode:
	@echo "Running VS Code integration tests..."
	./tests/run_vscode.sh

# Run only AI provider tests
test-providers:
	@echo "Running AI provider tests..."
	./tests/run_providers.sh

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Test with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test -bench=. ./...

# Run specific test
test-specific TEST:
	@echo "Running specific test: {{TEST}}"
	go test -run {{TEST}} ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f ./nixai
	rm -rf ./dist
	rm -f coverage.out coverage.html
	rm -f mcp.log

# Deep clean including modules cache
clean-all: clean
	@echo "Deep cleaning..."
	go clean -modcache
	rm -rf vendor/

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofumpt -w .

# Lint the code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Fix linting issues automatically
lint-fix:
	@echo "Fixing linting issues..."
	golangci-lint run --fix ./...

# Security check
security:
	@echo "Running security checks..."
	gosec ./...

# Check for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Vendor dependencies
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor

# Generate documentation
doc:
	@echo "Generating documentation..."
	go doc ./...

# Generate API documentation
doc-api:
	@echo "Generating API documentation..."
	godoc -http=:6060

# All-in-one command to build, test, and run
all: clean build test run

# Development workflow
dev: deps fmt lint test build

# CI workflow
ci: deps fmt lint test-coverage security vuln-check build

# Nix development environment
nix-develop:
	@echo "Entering Nix development shell..."
	nix develop

# Install nixai globally using Nix
nix-install:
	@echo "Installing nixai using Nix..."
	nix profile install .

# Build using Nix
nix-build:
	@echo "Building nixai using Nix..."
	nix build

# Test nixos configuration parsing
test-nixos-parse LOG_FILE:
	@echo "Testing NixOS log parsing with: {{LOG_FILE}}"
	./nixai parse --file {{LOG_FILE}}

# Test AI integration
test-ai PROVIDER:
	@echo "Testing AI integration with provider: {{PROVIDER}}"
	./nixai diagnose --provider {{PROVIDER}} --test

# Start MCP server in background
mcp-start:
	@echo "Starting MCP server in background..."
	nohup ./nixai mcp-server start > mcp.log 2>&1 &
	@echo "MCP server started, check mcp.log for output"

# Stop MCP server
mcp-stop:
	@echo "Stopping MCP server..."
	pkill -f "nixai mcp-server" || echo "No MCP server running"

# Check MCP server status
mcp-status:
	@echo "Checking MCP server status..."
	pgrep -f "nixai mcp-server" && echo "MCP server is running" || echo "MCP server is not running"

# View MCP server logs
mcp-logs:
	@echo "Viewing MCP server logs..."
	tail -f mcp.log

# Generate sample configurations
sample-configs:
	@echo "Generating sample configurations..."
	mkdir -p examples
	./nixai config --generate-samples --output examples/

# Validate configuration
validate-config:
	@echo "Validating configuration..."
	./nixai config --validate

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/godoc@latest

# Help command to list available commands
help:
	@echo "Available commands for nixai project:"
	@echo ""
	@echo "Building:"
	@echo "  build         - Build the application"
	@echo "  build-prod    - Build for production with optimizations"
	@echo "  build-all     - Build for multiple architectures"
	@echo "  nix-build     - Build using Nix"
	@echo ""
	@echo "Running:"
	@echo "  run           - Run the application"
	@echo "  run-args      - Run with specific arguments"
	@echo "  run-interactive - Run in interactive mode"
	@echo "  run-mcp       - Run MCP server"
	@echo "  run-debug     - Run with debug logging"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run tests"
	@echo "  test-all      - Run all tests (including Go tests, MCP tests, VS Code integration tests, provider tests)"
	@echo "  test-mcp      - Run only MCP tests"
	@echo "  test-vscode   - Run only VS Code integration tests"
	@echo "  test-providers - Run only AI provider tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-bench    - Run benchmark tests"
	@echo "  test-specific - Run specific test"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           - Format the code"
	@echo "  lint          - Lint the code"
	@echo "  lint-fix      - Fix linting issues automatically"
	@echo "  security      - Run security checks"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps          - Install dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  vendor        - Vendor dependencies"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean         - Clean build artifacts"
	@echo "  clean-all     - Deep clean including modules cache"
	@echo ""
	@echo "Documentation:"
	@echo "  doc           - Generate documentation"
	@echo "  doc-api       - Generate API documentation"
	@echo ""
	@echo "Workflows:"
	@echo "  all           - Clean, build, test, and run"
	@echo "  dev           - Development workflow"
	@echo "  ci            - CI workflow"
	@echo ""
	@echo "Nix Integration:"
	@echo "  nix-develop   - Enter Nix development shell"
	@echo "  nix-install   - Install nixai globally using Nix"
	@echo ""
	@echo "MCP Server:"
	@echo "  mcp-start     - Start MCP server in background"
	@echo "  mcp-stop      - Stop MCP server"
	@echo "  mcp-status    - Check MCP server status"
	@echo "  mcp-logs      - View MCP server logs"
	@echo ""
	@echo "Testing Features:"
	@echo "  test-nixos-parse - Test NixOS configuration parsing"
	@echo "  test-ai       - Test AI integration"
	@echo ""
	@echo "Configuration:"
	@echo "  sample-configs - Generate sample configurations"
	@echo "  validate-config - Validate configuration"
	@echo ""
	@echo "Setup:"
	@echo "  setup-dev     - Setup development environment"
	@echo "  doc     - Generate documentation"
	@echo "  all     - Clean, build, test, and run"
	@echo "  help    - Show this help message"