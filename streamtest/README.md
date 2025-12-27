# streamtest

Test video streamer for bs-image-stream-server. This tool decodes an MP4 video file and writes frames to `/tmp/output.jpg` at the video's native frame rate, simulating the input that bs-image-stream-server expects.

## Requirements

- **ffmpeg** and **ffprobe** must be installed

Install on Ubuntu/Debian:
```bash
sudo apt-get install ffmpeg
```

## Building

```bash
# Build the binary to ../bin/streamtest
make build
```

## Usage

```bash
# Basic usage - play video once
../bin/streamtest -file video.mp4

# Loop video continuously
../bin/streamtest -file video.mp4 -loop

# Verbose output showing every frame
../bin/streamtest -file video.mp4 -verbose

# Combine options
../bin/streamtest -file video.mp4 -loop -verbose
```

## Options

- `-file <path>` - Path to MP4 video file (required)
- `-loop` - Loop video playback continuously
- `-verbose` - Print detailed information for every frame

## Testing with bs-image-stream-server

1. Start the image stream server in one terminal:
```bash
cd /workspace
./bin/image-stream-server -file /tmp/output.jpg
```

2. Start the video streamer in another terminal:
```bash
cd /workspace/streamtest
../bin/streamtest -file your-video.mp4 -loop
```

3. Open your browser to http://localhost:8080 to see the video stream

## How It Works

1. Uses **ffprobe** to analyze the video file and extract metadata (resolution, frame rate, duration)
2. Uses **ffmpeg** to decode the video into JPEG frames via a pipe
3. Reads JPEG frames from the ffmpeg output stream by detecting JPEG markers (SOI/EOI)
4. Writes each frame to `/tmp/output.jpg` with timing to match the video's frame rate
5. Supports looping for continuous testing

## Output

The tool displays real-time statistics:
```
Video: test.mp4
Resolution: 1920x1080
Frame Rate: 30.00 fps
Duration: 10.50 seconds

Starting video stream... (Ctrl+C to stop)
Frames: 240 | FPS: 29.98
```

With `-verbose`, it shows per-frame details:
```
Frame  315 | Size:    42 KB | Target: 30.00 fps | Actual: 29.97 fps
```
