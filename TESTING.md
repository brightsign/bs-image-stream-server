# Testing bs-image-stream-server

This guide shows how to test the image stream server using the included `streamtest` tool.

## Overview

The testing setup consists of two programs:

1. **image-stream-server** - Monitors `/tmp/output.jpg` and serves it via HTTP at 30 FPS
2. **streamtest** - Decodes an MP4 video and writes frames to `/tmp/output.jpg`

## Prerequisites

Install ffmpeg (required by streamtest):
```bash
sudo apt-get update
sudo apt-get install ffmpeg
```

## Building

Build both programs:
```bash
# Build everything
make build
make build-streamtest

# Or build individually
make build           # Builds image-stream-server
make build-streamtest # Builds streamtest
```

Binaries are created in `./bin/`:
- `./bin/image-stream-server`
- `./bin/streamtest`

## Testing Workflow

### Terminal 1: Start the Image Stream Server

```bash
./bin/image-stream-server -file /tmp/output.jpg
```

You should see:
```
BrightSign Image Stream Server
Port: 8080
File: /tmp/output.jpg
Refresh Rate: 30 FPS

Server started at http://localhost:8080
Press Ctrl+C to stop
```

### Terminal 2: Stream a Video

```bash
# Play video once
./bin/streamtest -file your-video.mp4

# Loop video continuously
./bin/streamtest -file your-video.mp4 -loop

# Verbose output
./bin/streamtest -file your-video.mp4 -loop -verbose
```

You should see:
```
Video: your-video.mp4
Resolution: 1920x1080
Frame Rate: 30.00 fps
Duration: 10.50 seconds
Output: /tmp/output.jpg

Starting video stream... (Ctrl+C to stop)
Frames: 240 | FPS: 29.98
```

### Terminal 3 (or Browser): View the Stream

**Option 1: Browser**
Open http://localhost:8080 in your web browser

**Option 2: curl (test)**
```bash
# Get the current frame
curl http://localhost:8080/image -o test.jpg

# Check server health
curl http://localhost:8080/health
```

## Performance Testing

### Monitor Frame Rate

In verbose mode, streamtest shows real-time performance:
```bash
./bin/streamtest -file video.mp4 -verbose
```

Output:
```
Frame  315 | Size:    42 KB | Target: 30.00 fps | Actual: 29.97 fps
```

### Check Server Performance

The server logs show cache update frequency:
```
2025/12/27 12:00:00 Watching for changes to /tmp/output.jpg
```

### Load Testing

Test multiple concurrent clients:
```bash
# Terminal 1: Start server
./bin/image-stream-server

# Terminal 2: Stream video
./bin/streamtest -file video.mp4 -loop

# Terminal 3: Load test (requires while loop)
make load-test

# Or run multiple curl clients
for i in {1..10}; do
  while true; do curl -s http://localhost:8080/image > /dev/null; sleep 0.033; done &
done
```

## Test Scenarios

### 1. Single Client Playback
- Start server
- Start streamtest with loop
- Open browser to http://localhost:8080
- Verify smooth playback

### 2. Multiple Clients
- Start server
- Start streamtest with loop
- Open multiple browser tabs/windows
- Verify all clients show synchronized frames

### 3. Different Frame Rates
Test videos with different frame rates:
```bash
./bin/streamtest -file 24fps-video.mp4 -loop
./bin/streamtest -file 30fps-video.mp4 -loop
./bin/streamtest -file 60fps-video.mp4 -loop
```

### 4. High Resolution
Test with high-resolution videos:
```bash
./bin/streamtest -file 4k-video.mp4 -loop -verbose
```

Monitor frame size and performance.

### 5. Continuous Operation
Test server stability over extended periods:
```bash
# Leave running for hours/days
./bin/streamtest -file video.mp4 -loop
```

Monitor memory usage with `top` or `htop`.

## Troubleshooting

### streamtest: "ffmpeg not found"
```bash
sudo apt-get install ffmpeg
```

### streamtest: "Video file not found"
Verify the video file path:
```bash
ls -lh your-video.mp4
```

### Server: "Server not available"
Ensure the server is running:
```bash
curl http://localhost:8080/health
```

### Browser: Image not updating
1. Check streamtest is writing frames (watch the counter)
2. Check browser console for errors
3. Hard refresh the page (Ctrl+Shift+R)

### Performance Issues
1. Check system resources: `top` or `htop`
2. Use smaller resolution videos for testing
3. Monitor streamtest FPS output - should match video frame rate
4. Verify `/tmp` has sufficient space: `df -h /tmp`

## Expected Performance

On a typical embedded Linux system:

- **Frame Rate**: 30 FPS sustained
- **Latency**: < 100ms from file write to stream delivery
- **CPU Usage**: < 10% per client (depends on resolution)
- **Memory**: ~10MB base + ~5MB per concurrent client
- **Bandwidth**: Depends on JPEG quality and resolution
  - 720p: ~200-500 KB/frame = 6-15 MB/s @ 30fps
  - 1080p: ~400-800 KB/frame = 12-24 MB/s @ 30fps

## Advanced Testing

### Measure Latency

Use the included latency measurement tool:
```bash
cd cmd/measure_latency
go run measure_latency.go
```

### Custom Test Videos

Create test videos with specific characteristics:
```bash
# Generate 30fps test pattern
ffmpeg -f lavfi -i testsrc=duration=10:size=1920x1080:rate=30 \
  -c:v libx264 -pix_fmt yuv420p test-30fps.mp4

# Generate 60fps test pattern
ffmpeg -f lavfi -i testsrc=duration=10:size=1920x1080:rate=60 \
  -c:v libx264 -pix_fmt yuv420p test-60fps.mp4
```

## Cleanup

Stop all processes:
```bash
# Press Ctrl+C in each terminal

# Clean up test files
rm -f /tmp/output.jpg
rm -f test.jpg

# Clean binaries
make clean
cd streamtest && make clean
```
