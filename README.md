# BS Frame Monitor

A high-performance embedded Linux web server written in Go that provides real-time image streaming capabilities for monitoring video output from digital signage and machine vision applications.

## Overview

The BS Frame Monitor continuously monitors a local image file and serves it via HTTP at 30 FPS to web browsers, making it ideal for real-time visual monitoring of digital signage displays, camera feeds, or any application that generates periodic image updates.

### Key Features

- **High-Performance Streaming**: 30 FPS image updates with minimal CPU overhead
- **Embedded System Optimized**: Designed for resource-constrained Linux environments
- **Professional UI**: BrightSign-branded web interface with modern gradient styling
- **Efficient Change Detection**: Uses file modification time instead of content hashing
- **Zero Dependencies**: Built with Go standard library only
- **Cross-Platform**: Supports ARM, ARM64, and x86_64 architectures
- **ETag Support**: Bandwidth optimization with HTTP 304 Not Modified responses
- **Graceful Shutdown**: Clean shutdown with signal handling (SIGINT/SIGTERM)

### Architecture

- **Web Server**: HTTP server serving static HTML and real-time image endpoints
- **File Monitor**: Efficient 30 FPS polling with modification time detection
- **Image Cache**: Thread-safe memory caching with ETag support
- **Professional Interface**: Responsive web UI with BrightSign branding

## Quick Start

### Prerequisites

- Go 1.21.5 or later
- Linux, macOS, or Windows with WSL
- A source that generates images to `/tmp/output.jpg`

### Build and Run

```bash
# Clone the repository
git clone <repository-url>
cd bs-frame-monitor

# Build the application
make build

# Run with default settings
make run

# Run with debug logging
make run-debug

# Run with custom settings
make run-custom  # runs on port 8080 with custom file path
```

### Quick Build Options

```bash
# Show all available make targets
make help

# Build and run in one command
make run

# Development cycle (format, vet, test, build)
make dev-cycle

# Clean all build artifacts
make clean

# Initialize go modules
make deps
```

## Project Structure

```
bs-frame-monitor/
├── CLAUDE.md                      # Claude Code guidance and implementation plan
├── Makefile                       # Build automation and tasks
├── README.md                      # This file
├── go.mod                         # Go module definition
├── main.go                        # Application entry point
├── internal/
│   ├── cache/
│   │   ├── image_cache.go         # Thread-safe image caching
│   │   └── image_cache_test.go    # Cache unit tests
│   ├── monitor/
│   │   ├── file_monitor.go        # 30 FPS file monitoring
│   │   └── file_monitor_test.go   # Monitor unit tests
│   ├── server/
│   │   ├── server.go              # HTTP server setup
│   │   ├── handlers.go            # HTTP request handlers
│   │   ├── handlers_test.go       # Handler unit tests
│   │   └── static/
│   │       └── index.html         # BrightSign-branded web interface
│   └── testutil/
│       └── image_generator.go     # Test image generation utilities
├── integration_test.go            # End-to-end integration tests
├── load_test.go                   # Performance load testing
└── test-plan.md                   # Comprehensive test plan
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Serves the BrightSign-branded HTML interface |
| `/image` | GET | Returns the current JPEG image with ETag support |
| `/health` | GET | Health check endpoint returning JSON status |
| `/images/*` | GET | Static file server for assets (logos, etc.) |

## Command Line Options

```bash
./bs-frame-monitor [options]

Options:
  -port int
        HTTP server port (default 8080)
  -file string
        Path to image file to monitor (default "/tmp/output.jpg")
  -debug
        Enable debug logging
```

## Building for Embedded Targets

All build targets automatically disable CGO for static binary compilation.

### Build for all platforms
```bash
make build-embedded
```

### Build for specific platforms
```bash
# Linux x86_64
make build-linux

# ARM (Raspberry Pi 3, 32-bit)
make build-arm

# ARM64 (Raspberry Pi 4, 64-bit)
make build-arm64

# BrightSign/Rockchip ARM64 player
make player
```

### Build output locations
All binaries are placed in the `cmd/` directory:
- `cmd/image-stream-server-amd64` - Linux x86_64
- `cmd/image-stream-server-arm` - ARM 32-bit
- `cmd/image-stream-server-arm64` - ARM 64-bit
- `cmd/image-stream-server-player` - BrightSign player

## Development

### Run tests
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests only
make test ARGS="./integration_test.go"

# Run load tests only
make test ARGS="./load_test.go"
```

### Code quality
```bash
# Format code
make fmt

# Vet code
make vet

# Run linter (requires golangci-lint)
make lint

# Full development cycle
make dev-cycle
```

### Load testing
```bash
# Simple load test against localhost:8080
make load-test
```

## Deployment

### Systemd Service

Create `/etc/systemd/system/bs-frame-monitor.service`:

```ini
[Unit]
Description=BrightSign Frame Monitor
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/bs-frame-monitor
ExecStart=/opt/bs-frame-monitor/bs-frame-monitor -port 8080 -file /tmp/output.jpg
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable bs-frame-monitor
sudo systemctl start bs-frame-monitor
```

### Static Binary Deployment

The application compiles to a single static binary with no dependencies:

```bash
# Build static binary for deployment
make build-linux

# Copy to target system
scp cmd/image-stream-server-amd64 user@target:/opt/bs-frame-monitor/
```

## Performance Optimization

The server is optimized for embedded systems with:

- **Memory efficiency**: Reuses buffers, minimizes allocations
- **CPU efficiency**: Only reads files when modification time changes
- **Network efficiency**: ETag support reduces bandwidth usage
- **Concurrent handling**: Thread-safe operations throughout

## Use Cases

- **Digital Signage Monitoring**: Real-time view of display output
- **Camera Feed Streaming**: Live monitoring of security or industrial cameras
- **Automated Testing**: Visual verification of display content in CI/CD pipelines
- **Remote Troubleshooting**: Quick visual assessment of remote display issues
- **Machine Vision**: Monitor output from image processing applications

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]