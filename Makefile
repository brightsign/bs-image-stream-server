# BS Frame Monitor Makefile
# Build configuration for embedded Linux image streaming server

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=image-stream-server
BINARY_DIR=cmd
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
BINARY_UNIX=$(BINARY_DIR)/$(BINARY_NAME)-amd64
BINARY_ARM=$(BINARY_DIR)/$(BINARY_NAME)-arm
BINARY_ARM64=$(BINARY_DIR)/$(BINARY_NAME)-arm64

# Build flags
LDFLAGS=-ldflags "-w -s"
BUILD_FLAGS=-a -installsuffix cgo

# Default target
.PHONY: all
all: test build

# Build the binary for current platform
.PHONY: build
build:
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) .

# Build for Linux (x86_64)
.PHONY: build-linux
build-linux:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_UNIX) .

# Build for ARM (Raspberry Pi 3, etc.)
.PHONY: build-arm
build-arm:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_ARM) .

# Build for ARM64 (Raspberry Pi 4, etc.)
.PHONY: build-arm64
build-arm64:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_ARM64) .

# Build for Rockchip ARM64 player
.PHONY: player
player:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-player .

# Build all embedded targets
.PHONY: build-embedded
build-embedded: build-linux build-arm build-arm64

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_ARM)
	rm -f $(BINARY_ARM64)
	rm -f image-stream-server
	rm -f image-stream-server-amd64
	rm -f image-stream-server-arm
	rm -f image-stream-server-arm64
	rm -f $(BINARY_DIR)/$(BINARY_NAME)-player
	rm -f coverage.out
	rm -f coverage.html

# Run the application locally
.PHONY: run
run: build
	./$(BINARY_PATH)

# Run with debug logging
.PHONY: run-debug
run-debug: build
	./$(BINARY_PATH) -debug

# Run with custom settings
.PHONY: run-custom
run-custom: build
	./$(BINARY_PATH) -port 8080 -file /tmp/custom.jpg -refresh 60

# Initialize Go modules
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format Go code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Lint Go code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Vet Go code
.PHONY: vet
vet:
	$(GOCMD) vet ./...

# Development cycle: format, vet, test, build
.PHONY: dev-cycle
dev-cycle: fmt vet test build

# Interactive development shell in Docker container
.PHONY: dev
dev: docker-build
	docker run -it --rm -v $(PWD):/workspace claude-dev-env

# Docker targets
.PHONY: docker-build
docker-build:
	docker build -t claude-dev-env .

.PHONY: docker-run
docker-run:
	docker run -it --rm -v $(PWD):/workspace claude-dev-env

.PHONY: docker-claude
docker-claude:
	docker run -it --rm -v $(PWD):/workspace claude-dev-env bash -c "cd /workspace && claude --dangerously-skip-permissions"

# Load testing (requires curl)
.PHONY: load-test
load-test:
	@echo "Starting load test (Ctrl+C to stop)..."
	@while true; do curl -s http://localhost:8080/image > /dev/null; sleep 0.033; done

# Create directory structure for new project
.PHONY: init-project
init-project:
	$(GOMOD) init bs-frame-monitor
	mkdir -p cmd
	mkdir -p internal/server
	mkdir -p internal/monitor
	mkdir -p internal/cache
	mkdir -p web
	mkdir -p pkg
	mkdir -p test
	touch cmd/main.go
	touch internal/server/server.go
	touch internal/server/handlers.go
	touch internal/monitor/file_monitor.go
	touch internal/cache/image_cache.go
	touch web/index.html

# Install development dependencies
.PHONY: install-dev-deps
install-dev-deps:
	$(GOGET) -u golang.org/x/tools/cmd/goimports
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint


# Show help
.PHONY: help
help:
	@echo "BS Frame Monitor - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build           - Build binary for current platform"
	@echo "  build-linux     - Build for Linux x86_64"
	@echo "  build-arm       - Build for ARM (Raspberry Pi 3)"
	@echo "  build-arm64     - Build for ARM64 (Raspberry Pi 4)"
	@echo "  player          - Build for Rockchip ARM64 player"
	@echo "  build-embedded  - Build all embedded targets"
	@echo ""
	@echo "Development targets:"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  fmt             - Format Go code"
	@echo "  lint            - Lint Go code (requires golangci-lint)"
	@echo "  vet             - Vet Go code"
	@echo "  dev-cycle       - Run full development cycle"
	@echo "  dev             - Start interactive Docker development shell"
	@echo "  deps            - Download and tidy Go modules"
	@echo ""
	@echo "Run targets:"
	@echo "  run             - Build and run locally"
	@echo "  run-debug       - Run with debug logging"
	@echo "  run-custom      - Run with custom settings"
	@echo "  load-test       - Run load test against localhost:8080"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build    - Build development Docker image"
	@echo "  docker-run      - Run interactive Docker container"
	@echo "  docker-claude   - Run Claude Code in Docker container"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean           - Clean build artifacts"
	@echo "  init-project    - Create project directory structure"
	@echo "  install-dev-deps - Install development dependencies"
	@echo "  systemd-service - Create systemd service file"
	@echo "  help            - Show this help message"