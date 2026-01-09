.PHONY: build install test test-verbose test-coverage test-race test-update clean fmt vet lint help

# Default target
.DEFAULT_GOAL := help

# Build the CLI tool
build:
	@echo "Building houp..."
	go build -o houp ./cmd/houp

# Build with verbose output
build-verbose:
	@echo "Building houp with verbose output..."
	go build -v ./...

# Install to GOPATH/bin
install:
	@echo "Installing houp..."
	go install ./cmd/houp

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./pkg/generator -cover

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Update golden files
test-update:
	@echo "Updating golden files..."
	go test ./pkg/generator -update

# Format all code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet for static analysis
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all linting checks
lint: fmt vet
	@echo "Running all linting checks..."
	go mod verify
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f houp

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the CLI tool"
	@echo "  make build-verbose  - Build with verbose output"
	@echo "  make install        - Install to GOPATH/bin"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make test-race      - Run tests with race detection"
	@echo "  make test-update    - Update golden files"
	@echo "  make fmt            - Format all code"
	@echo "  make vet            - Run go vet for static analysis"
	@echo "  make lint           - Run all linting checks (fmt, vet, mod verify)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make help           - Show this help message"
