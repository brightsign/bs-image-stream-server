# bs-image-stream-server

A high-performance embedded Linux web server written in Go that provides real-time image streaming capabilities for monitoring video output from digital signage and machine vision applications.

## Overview

The bs-image-stream-server continuously monitors a local image file and serves it via HTTP at 30 FPS to web browsers, making it ideal for real-time visual monitoring of digital signage displays, camera feeds, or any application that generates periodic image updates.

## How to Use

### Basic Usage

1. **Start the server** with your image file path:
   ```bash
   ./bs-image-stream-server -file /path/to/your/image.jpg
   ```

2. **View the live stream** in your browser:
   - Navigate to `http://localhost:8080/` or `http://localhost:8080/video`
   - The image will automatically refresh at 30 FPS

3. **Direct image access** for integration:
   - Access `http://localhost:8080/image` to get the raw JPEG data
   - Supports ETag headers for efficient caching

### Common Use Cases

- **Monitor BrightSign player output**: Point to the screenshot file your player generates
- **Security camera feeds**: Monitor JPEG files updated by IP cameras
- **Image processing pipelines**: View the output of computer vision applications
- **Remote diagnostics**: Check display content from anywhere on your network

### Key Features

- **High-Performance Streaming**: 30 FPS image updates with minimal CPU overhead
- **Embedded System Optimized**: Designed for resource-constrained Linux environments
- **Professional UI**: BrightSign-branded web interface with modern gradient styling
- **Efficient Change Detection**: Uses file modification time instead of content hashing
- **Zero Dependencies**: Built with Go standard library only, with assets embedded in binary
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
cd bs-image-stream-server

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

## Practical Examples

### Example 1: Monitor BrightSign Player Output

If your BrightSign player writes screenshots to `/tmp/display.jpg`:

```bash
# Start the monitor
./bs-image-stream-server -port 8080 -file /tmp/display.jpg

# View with web interface
# http://player-ip:8080/

# View raw MJPEG stream
# http://player-ip:8080/video
```

### Example 2: Security Camera Integration

Monitor a camera that saves snapshots:

```bash
# Monitor camera snapshot
./bs-image-stream-server -port 8080 -file /var/cameras/front-door.jpg

# View from anywhere on your network
# http://server-ip:8080/
```

### Example 3: Custom Application Integration

Fetch the image programmatically:

```bash
# Get current image with curl
curl -H "If-None-Match: \"12345-6789\"" http://localhost:8080/image

# Response includes ETag header for caching
# Returns 304 if image hasn't changed
```

### Example 4: Video Recording and Streaming

Record with ffmpeg or use in video software:

```bash
# Record with ffmpeg (use format=mjpeg for compatibility)
ffmpeg -i http://localhost:8080/video?format=mjpeg -c copy output.avi

# Record as MP4 with re-encoding
ffmpeg -i http://localhost:8080/video?format=mjpeg -c:v libx264 -r 30 output.mp4

# View in VLC media player (both formats work)
vlc http://localhost:8080/video
vlc http://localhost:8080/video?format=mjpeg

# Embed in HTML (use default multipart format)
<img src="http://server:8080/video" />

# Use with streaming software like OBS
# Add "Media Source" with URL: http://server:8080/video?format=mjpeg
```

### Example 5: Load Balancer Configuration

Use the health endpoint for monitoring:

```nginx
upstream bs_monitor {
    server localhost:8080;
    health_check uri=/health;
}
```

## Project Structure

```
bs-image-stream-server/
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

### Web Interface Endpoints

| Endpoint | Method | Description | Use Case |
|----------|--------|-------------|----------|
| `/` | GET | HTML viewing interface with BrightSign branding | General monitoring and viewing with web interface |
| `/video` | GET | Multipart MJPEG stream (default) | Browser viewing without HTML wrapper |
| `/video?format=mjpeg` | GET | Raw MJPEG stream for recording | ffmpeg recording and video software |

- `/` provides a branded web interface with JavaScript-based 30 FPS refresh
- `/video` provides a multipart MJPEG stream that browsers can display directly
- `/video?format=mjpeg` provides a raw MJPEG stream compatible with video recording tools

### API Endpoints

| Endpoint | Method | Description | Use Case |
|----------|--------|-------------|----------|
| `/image` | GET | Raw JPEG image with caching headers | Direct image access for custom applications |
| `/health` | GET | JSON health status | Monitoring and load balancer health checks |
| `/images/*` | GET | Static assets (logos, etc.) | Internal use by the web interface |

### Endpoint Details

#### `/` - Web Viewing Interface
- **Purpose**: Human-friendly web interface for viewing the live image stream
- **Features**: 
  - Auto-refreshes at 30 FPS using JavaScript
  - Clean, branded interface with purple frame
  - BrightSign logo and professional styling
  - Works in any modern browser
- **When to use**: When you need a polished interface for monitoring

#### `/video` - MJPEG Streaming
- **Purpose**: Live video streaming in different formats
- **Default behavior**: Multipart MJPEG stream for browsers
- **Features**:
  - Native browser video streaming at 30 FPS
  - No HTML wrapper or JavaScript required
  - Lower bandwidth than JavaScript refresh approach
- **Format options**:
  - `/video` - Multipart stream (browser-friendly)
  - `/video?format=mjpeg` - Raw MJPEG (ffmpeg-compatible)
- **When to use**:
  - Browser viewing without web interface
  - Video recording with ffmpeg
  - Embedding in other applications
  - Video streaming software integration

#### `/image` - Direct Image Access
- **Purpose**: Programmatic access to the current image
- **Features**:
  - Returns raw JPEG data
  - Includes ETag header for efficient caching
  - Returns 304 Not Modified if image hasn't changed
  - Ideal for custom applications or embedding
- **When to use**: 
  - Building custom viewing applications
  - Integrating with other systems
  - Creating image processing pipelines
  - When you need efficient bandwidth usage with ETag support

#### `/health` - System Health Check
- **Purpose**: Monitor server status
- **Response format**: 
  ```json
  {
    "status": "ok",
    "timestamp": "2024-01-15T10:30:00Z"
  }
  ```
- **Status values**:
  - `"ok"` - Server is running and has image data
  - `"no_image"` - Server is running but no image is available yet
- **When to use**: Load balancer health checks, monitoring systems

## Command Line Options

```bash
./bs-image-stream-server [options]

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

# Note: BrightSign/Rockchip ARM64 player is built automatically with 'make build'
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

Create `/etc/systemd/system/bs-image-stream-server.service`:

```ini
[Unit]
Description=BrightSign Image Stream Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/bs-image-stream-server
ExecStart=/opt/bs-image-stream-server/bs-image-stream-server -port 8080 -file /tmp/output.jpg
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable bs-image-stream-server
sudo systemctl start bs-image-stream-server
```

### Static Binary Deployment

The application compiles to a single self-contained static binary with no external dependencies or files:

```bash
# Build static binary for deployment
make build-linux

# Copy single binary to target system - no additional files needed
scp cmd/image-stream-server-amd64 user@target:/opt/bs-image-stream-server/

# The binary includes all assets (HTML, CSS, logo) embedded
# No need to copy additional files
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
- **Video Recording**: Capture live streams for analysis or archival

## Troubleshooting

### ffmpeg Recording Issues

If ffmpeg fails to record, try these solutions:

```bash
# Use the MJPEG format explicitly
ffmpeg -i http://server:8080/video?format=mjpeg -c copy output.avi

# If -c copy fails, re-encode the stream
ffmpeg -i http://server:8080/video?format=mjpeg -c:v libx264 -r 30 output.mp4

# For longer recordings, add duration limit
ffmpeg -i http://server:8080/video?format=mjpeg -t 60 -c copy output.avi

# If connection issues occur, increase buffer size
ffmpeg -i http://server:8080/video?format=mjpeg -buffer_size 32768 -c copy output.avi
```

### Browser Compatibility

- Use `/video` (default) for browser viewing
- Use `/video?format=mjpeg` for video recording tools
- Some older browsers may not support multipart streams

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]