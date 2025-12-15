# Makefile
.PHONY: build test lint clean install run

# Variables
BINARY_NAME=dev-cleaner
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Install locally
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Run the tool
run:
	go run . $(ARGS)

# All checks before commit
check: fmt vet test
	@echo "✅ All checks passed!"

# Quick scan (dev helper)
scan: build
	./$(BINARY_NAME) scan
