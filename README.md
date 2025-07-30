# bs-image-stream-server

A high-performance embedded Linux web server written in Go that provides real-time image streaming capabilities for monitoring video output from digital signage and machine vision applications.

## Overview

The bs-image-stream-server continuously monitors a local image file and serves it via HTTP at 30 FPS, making it ideal for real-time visual monitoring of digital signage displays, camera feeds, or any application that generates periodic image updates.

### How It Works

The server operates using a simple but effective architecture:

1. **File Monitoring**: Watches a specified image file (e.g., `/tmp/output.jpg`) using 33ms intervals (30 FPS)
2. **Change Detection**: Uses file modification time comparison to detect when the image updates
3. **Memory Caching**: Stores the current image in memory with ETag support for efficient serving
4. **Multi-Format Streaming**: Serves the image through multiple endpoints:
   - **Web Interface** (`/`): HTML page with JavaScript-based 30 FPS refresh
   - **MJPEG Stream** (`/video`): Direct multipart MJPEG stream for browsers and ffmpeg
   - **Raw Image** (`/image`): Direct JPEG access with HTTP caching headers
5. **Zero Dependencies**: Single binary with all assets embedded - no external files needed

## How to Use

### Basic Usage

1. **Start the server** with your image file path:
   ```bash
   ./bs-image-stream-server -file /path/to/your/image.jpg
   ```

2. **View the live stream** in your browser:
   - Navigate to `http://<player>:8080/` or `http://<player>:8080/video`
   - The image will automatically refresh at 30 FPS

3. **Direct image access** for integration:
   - Access `http://<player>:8080/image` to get the raw JPEG data
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
- Optional: [bscp](https://github.com/gherlein/bs-scp) for BrightSign deployment

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

# Build and install to BrightSign player
make install PLAYER=<player-hostname>

# Development cycle (format, vet, test, build)
make dev-cycle

# Clean all build artifacts
make clean

# Initialize go modules
make deps
```

## Usage

### Basic Viewing

Start the server and view the stream in your browser:

```bash
# Start monitoring an image file
./bs-image-stream-server -file /path/to/image.jpg -port 8080

# View in browser with web interface
open http://<player>:8080/

# View raw video stream (no HTML wrapper)
open http://<player>:8080/video
```

### Recording with ffmpeg

The `/video` endpoint provides an MJPEG stream that ffmpeg can record:

```bash
# Basic recording (copies stream as-is)
ffmpeg -i http://<player>:8080/video -c copy recording.mpeg

# Record for specific duration (60 seconds)
ffmpeg -i http://<player>:8080/video -t 60 -c copy recording.mpeg

# Convert to MP4 while recording
ffmpeg -i http://<player>:8080/video -c:v libx264 -r 30 recording.mp4

# High quality MP4 recording
ffmpeg -i http://<player>:8080/video -c:v libx264 -crf 18 -r 30 high_quality.mp4

# Record with custom frame rate
ffmpeg -i http://<player>:8080/video -r 25 -c:v libx264 output.mp4
```

### Streaming to Other Services

Use the video stream with streaming platforms:

```bash
# Stream to YouTube Live (requires stream key)
ffmpeg -i http://<player>:8080/video -c:v libx264 -b:v 2500k -r 30 \
  -f flv rtmp://a.rtmp.youtube.com/live2/YOUR_STREAM_KEY

# Stream to Twitch (requires stream key)
ffmpeg -i http://<player>:8080/video -c:v libx264 -b:v 2500k -r 30 \
  -f flv rtmp://live.twitch.tv/app/YOUR_STREAM_KEY

# Re-stream to local RTMP server
ffmpeg -i http://<player>:8080/video -c:v libx264 -f flv rtmp://<rtmp-server>/live/stream
```

### Integration Examples

Embed or integrate the stream in applications:

```bash
# View in VLC media player
vlc http://<player>:8080/video

# Use with OBS Studio
# Add "Media Source" → Input: http://<player>:8080/video

# Embed in HTML page
echo '<img src="http://<player>:8080/video" alt="Live Stream">' > viewer.html

# Use with curl for testing
curl -N http://<player>:8080/video > stream_test.mjpeg
```

### Monitoring and Health Checks

```bash
# Check server health
curl http://<player>:8080/health

# Get current image directly
curl http://<player>:8080/image > current_frame.jpg

# Monitor with cache-aware requests
curl -H "If-None-Match: \"1234567890-12345\"" http://<player>:8080/image
```

## Practical Examples

### Example 1: Deploy to BrightSign Player

Deploy and monitor a BrightSign player directly:

```bash
# Deploy to BrightSign player
make install PLAYER=my-brightsign-player

# SSH to player and start monitoring
ssh brightsign@my-brightsign-player
/tmp/bs-image-stream-server -file /tmp/screenshot.jpg -port 8080 &

# From your computer, record the player output
ffmpeg -i http://my-brightsign-player:8080/video -t 300 -c copy player_recording.mpeg

# View live stream in browser
open http://my-brightsign-player:8080/
```

### Example 2: Security Camera Integration

Monitor a camera that saves snapshots and stream to YouTube:

```bash
# Start monitoring camera
./bs-image-stream-server -port 8080 -file /var/cameras/front-door.jpg

# Stream live to YouTube (requires stream key)
ffmpeg -i http://<player>:8080/video -c:v libx264 -b:v 1500k -r 25 \
  -f flv rtmp://a.rtmp.youtube.com/live2/YOUR_STREAM_KEY
```

### Example 3: Automated Monitoring and Recording

Set up automated recording with rotation:

```bash
# Start the server
./bs-image-stream-server -file /tmp/output.jpg &

# Record 1-hour segments with timestamps
while true; do
  timestamp=$(date +%Y%m%d_%H%M%S)
  ffmpeg -i http://<player>:8080/video -t 3600 -c copy "recording_${timestamp}.mpeg"
  sleep 10  # Brief pause between recordings
done
```

### Example 4: Load Balancer Configuration

Use the health endpoint for monitoring:

```nginx
upstream bs_monitor {
    server <player>:8080;
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
| `/video` | GET | Multipart MJPEG stream | Browser viewing and ffmpeg recording |

- `/` provides a branded web interface with JavaScript-based 30 FPS refresh
- `/video` provides an MJPEG stream that works with both browsers and recording tools

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
- **Purpose**: Live video streaming compatible with browsers and recording tools
- **Features**:
  - Native browser video streaming at 30 FPS
  - Direct ffmpeg recording compatibility (confirmed working)
  - No HTML wrapper or JavaScript required
  - Lower bandwidth than JavaScript refresh approach
  - Works with VLC, OBS, and other video software
- **When to use**:
  - Browser viewing without web interface
  - Video recording with ffmpeg (`ffmpeg -i http://<player>/video -c copy output.mpeg`)
  - Embedding in other applications
  - Streaming to platforms (YouTube, Twitch, etc.)

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

### BrightSign Player Deployment

Deploy directly to BrightSign players using the `make install` target:

```bash
# Install to default player hostname 'brightsign'
make install

# Install to specific player
make install PLAYER=my-player-name
make install PLAYER=192.168.1.100

# Set default player in environment
export PLAYER=my-brightsign-device
make install
```

**Requirements**: 
- [bscp](https://github.com/gherlein/bs-scp) must be installed
- Network access to the BrightSign player
- Player must have SSH enabled

**Installation process**:
1. Builds the ARM64 player binary
2. Copies binary to player's `/tmp/bs-image-stream-server`
3. Provides instructions for starting the server on the player

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
# Simple load test against running server
make load-test
```

## Deployment

### BrightSign Player Deployment

Use the built-in deployment system:

```bash
# Quick deployment to BrightSign player
make install PLAYER=<player-hostname>

# Then SSH to the player to start the service
ssh brightsign@<player-hostname>
/tmp/bs-image-stream-server -file /tmp/screenshot.jpg -port 8080
```

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

### ffmpeg Recording

The `/video` endpoint works directly with ffmpeg without requiring format parameters:

```bash
# Basic recording (confirmed working)
ffmpeg -i http://<player>:8080/video -c copy recording.mpeg

# If stream format issues occur, try re-encoding
ffmpeg -i http://<player>:8080/video -c:v libx264 -r 30 recording.mp4

# For longer recordings, add duration limit
ffmpeg -i http://<player>:8080/video -t 3600 -c copy recording.mpeg

# If connection issues occur, increase buffer size
ffmpeg -i http://<player>:8080/video -buffer_size 32768 -c copy recording.avi
```

**Expected behavior**: ffmpeg will detect the stream as `mpjpeg` format and achieve 25-30 FPS recording. The "Packet corrupt" warning at the end is normal when the stream ends.

### Stream Analysis

ffmpeg should show output similar to:
```
Input #0, mpjpeg, from 'http://<player>:8080/video':
  Duration: N/A, bitrate: N/A
  Stream #0:0: Video: mjpeg (Baseline), yuvj420p, 640x480, 25 tbr, 25 tbn
```

### Browser Compatibility

- **Web Interface** (`/`): Works in all modern browsers
- **Video Stream** (`/video`): Works in browsers that support multipart MJPEG
- **Direct Image** (`/image`): Universal compatibility with HTTP caching

### Common Issues

1. **No image displayed**: Check that the source file exists and is being updated
2. **Stream won't start**: Verify the file path and permissions
3. **ffmpeg connection refused**: Ensure the server is running and accessible
4. **Low frame rate**: Source file may not be updating at 30 FPS

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]