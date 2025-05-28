# Justfile for nixai project

# Justfile for nixai project
set shell := ["bash"]

# Build the application
build:
	go build -o ./nixai ./cmd/nixai/main.go

# Run the application
run: build
	./nixai

# Test the application
test:
	go test ./...

# Clean build artifacts
clean:
	go clean
	rm -f ./nixai

# Format the code
fmt:
	go fmt ./...

# Lint the code
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod tidy

# Generate documentation
doc:
	go doc ./...

# All-in-one command to build, test, and run
all: clean build test run

# Help command to list available commands
help:
	@echo "Available commands:"
	@echo "  build   - Build the application"
	@echo "  run     - Run the application"
	@echo "  test    - Test the application"
	@echo "  clean   - Clean build artifacts"
	@echo "  fmt     - Format the code"
	@echo "  lint    - Lint the code"
	@echo "  deps    - Install dependencies"
	@echo "  doc     - Generate documentation"
	@echo "  all     - Clean, build, test, and run"
	@echo "  help    - Show this help message"