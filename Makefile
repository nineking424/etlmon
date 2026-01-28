# etlmon Makefile

# Variables
BINARY_NAME=etlmon
VERSION=0.1.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
GO=go

# Default target
.PHONY: all
all: build

# Build binary
.PHONY: build
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/etlmon

# Build without CGO (for distribution)
.PHONY: build-static
build-static:
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/etlmon

# Build for Linux (cross-compile)
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/etlmon

# Build for all platforms
.PHONY: build-all
build-all: build-static build-linux

# Run tests
.PHONY: test
test:
	$(GO) test -v ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GO) test -race ./...

# Run tests with coverage
.PHONY: test-cover
test-cover:
	$(GO) test -cover ./...

# Generate coverage report
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GO) mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux-amd64
	rm -f coverage.out coverage.html
	rm -f *.db *.db-shm *.db-wal

# Run the application (requires config)
.PHONY: run
run: build
	./$(BINARY_NAME) --config configs/config.yaml

# Install to GOPATH/bin
.PHONY: install
install:
	$(GO) install $(LDFLAGS) ./cmd/etlmon

# Development: watch and rebuild (requires entr)
.PHONY: dev
dev:
	find . -name '*.go' | entr -r make run

# Comprehensive verification (build, test, coverage, race, vet)
.PHONY: verify
verify:
	@echo "Running build check..."
	$(GO) build ./...
	@echo "Build check passed."
	@echo ""
	@echo "Running tests with verbose output..."
	$(GO) test ./... -v
	@echo ""
	@echo "Generating coverage profile..."
	$(GO) test ./... -coverprofile=coverage.out
	@echo ""
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo ""
	@echo "Running race detector on TUI..."
	$(GO) test -race ./internal/tui/...
	@echo "Race detection passed."
	@echo ""
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Go vet passed."
	@echo ""
	@echo "Verification successful!"

# Show help
.PHONY: help
help:
	@echo "etlmon Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build binary"
	@echo "  make build-static - Build without CGO (for distribution)"
	@echo "  make build-linux  - Cross-compile for Linux"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make test         - Run tests"
	@echo "  make test-race    - Run tests with race detection"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make coverage     - Generate HTML coverage report"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make tidy         - Tidy dependencies"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make run          - Build and run with default config"
	@echo "  make install      - Install to GOPATH/bin"
	@echo "  make verify       - Run comprehensive verification"
	@echo "  make help         - Show this help"
