package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	outputPath = "/tmp/output.jpg"
)

// VideoInfo holds metadata about the video file
type VideoInfo struct {
	FrameRate float64
	Duration  float64
	Width     int
	Height    int
}

// FFProbeOutput structure for parsing ffprobe JSON
type FFProbeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		RFrameRate string `json:"r_frame_rate"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
	} `json:"streams"`
}

func main() {
	var (
		videoFile = flag.String("file", "", "Path to MP4 video file")
		loop      = flag.Bool("loop", false, "Loop video playback")
		verbose   = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	if *videoFile == "" {
		fmt.Println("Usage: streamtest -file <video.mp4> [-loop] [-verbose]")
		fmt.Println("\nOptions:")
		fmt.Println("  -file string     Path to MP4 video file (required)")
		fmt.Println("  -loop            Loop video playback continuously")
		fmt.Println("  -verbose         Print detailed frame information")
		os.Exit(1)
	}

	// Check if video file exists
	if _, err := os.Stat(*videoFile); os.IsNotExist(err) {
		log.Fatalf("Video file not found: %s", *videoFile)
	}

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("ffmpeg not found in PATH. Please install ffmpeg.")
	}

	// Get video information
	info, err := getVideoInfo(*videoFile)
	if err != nil {
		log.Fatalf("Failed to get video info: %v", err)
	}

	fmt.Printf("Video: %s\n", *videoFile)
	fmt.Printf("Resolution: %dx%d\n", info.Width, info.Height)
	fmt.Printf("Frame Rate: %.2f fps\n", info.FrameRate)
	fmt.Printf("Duration: %.2f seconds\n", info.Duration)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("\nStarting video stream... (Ctrl+C to stop)")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	stopChan := make(chan bool)
	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down...")
		close(stopChan)
	}()

	// Stream the video
	frameCount := 0
	for {
		count, err := streamVideo(*videoFile, info, *verbose, stopChan)
		frameCount += count

		if err != nil {
			if err == io.EOF {
				fmt.Printf("\nVideo ended after %d frames\n", frameCount)
			} else {
				log.Printf("Stream error: %v", err)
			}
		}

		select {
		case <-stopChan:
			fmt.Printf("Total frames streamed: %d\n", frameCount)
			return
		default:
			if !*loop {
				fmt.Printf("Total frames streamed: %d\n", frameCount)
				return
			}
			fmt.Println("\nLooping video...")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func getVideoInfo(videoFile string) (*VideoInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		videoFile)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %v", err)
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	// Find video stream
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			// Parse frame rate (format: "30/1" or "30000/1001")
			var num, den int
			fmt.Sscanf(stream.RFrameRate, "%d/%d", &num, &den)
			frameRate := float64(num) / float64(den)

			duration, _ := strconv.ParseFloat(stream.Duration, 64)

			return &VideoInfo{
				FrameRate: frameRate,
				Duration:  duration,
				Width:     stream.Width,
				Height:    stream.Height,
			}, nil
		}
	}

	return nil, fmt.Errorf("no video stream found")
}

func streamVideo(videoFile string, info *VideoInfo, verbose bool, stopChan chan bool) (int, error) {
	// Calculate frame delay for target frame rate
	frameDelay := time.Duration(float64(time.Second) / info.FrameRate)

	// Start ffmpeg to decode video and output JPEG frames
	cmd := exec.Command("ffmpeg",
		"-i", videoFile,
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", "3", // Quality (2-31, lower is better)
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	reader := bufio.NewReader(stdout)
	frameCount := 0
	startTime := time.Now()

	// JPEG markers
	const (
		jpegSOI = 0xFFD8 // Start of Image
		jpegEOI = 0xFFD9 // End of Image
	)

	for {
		select {
		case <-stopChan:
			cmd.Process.Kill()
			return frameCount, nil
		default:
		}

		frameStart := time.Now()

		// Read JPEG frame from stream
		frame, err := readJPEGFrame(reader)
		if err != nil {
			cmd.Wait()
			return frameCount, err
		}

		// Write frame to output file
		if err := os.WriteFile(outputPath, frame, 0644); err != nil {
			log.Printf("Failed to write frame: %v", err)
			continue
		}

		frameCount++
		elapsed := time.Since(startTime)
		actualFPS := float64(frameCount) / elapsed.Seconds()

		if verbose {
			fmt.Printf("\rFrame %4d | Size: %6d KB | Target: %.2f fps | Actual: %.2f fps",
				frameCount, len(frame)/1024, info.FrameRate, actualFPS)
		} else if frameCount%30 == 0 {
			fmt.Printf("\rFrames: %d | FPS: %.2f", frameCount, actualFPS)
		}

		// Sleep to maintain frame rate
		processingTime := time.Since(frameStart)
		sleepTime := frameDelay - processingTime
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}
}

func readJPEGFrame(reader *bufio.Reader) ([]byte, error) {
	var frame bytes.Buffer

	// Look for JPEG SOI marker (0xFF 0xD8)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == 0xFF {
			b2, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			if b2 == 0xD8 {
				frame.WriteByte(0xFF)
				frame.WriteByte(0xD8)
				break
			}
		}
	}

	// Read until JPEG EOI marker (0xFF 0xD9)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		frame.WriteByte(b)

		if b == 0xFF {
			b2, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			frame.WriteByte(b2)

			if b2 == 0xD9 {
				// Found end of JPEG
				return frame.Bytes(), nil
			}
		}
	}
}
