package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	testImagePath = "/tmp/output.jpg"
	serverURL     = "http://localhost:8080/video"
	imageWidth    = 640
	imageHeight   = 480
)

// ImageGenerator generates test images with timestamps
type ImageGenerator struct {
	frameCount int
}

func (g *ImageGenerator) GenerateImage(timestamp time.Time) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	// Fill with a color that changes over time
	g.frameCount++
	bgColor := color.RGBA{
		R: uint8((g.frameCount * 7) % 256),
		G: uint8((g.frameCount * 13) % 256),
		B: uint8((g.frameCount * 19) % 256),
		A: 255,
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw timestamp text
	timestampStr := timestamp.Format("15:04:05.000000")
	frameStr := fmt.Sprintf("Frame: %d", g.frameCount)

	addLabel(img, 20, 50, timestampStr)
	addLabel(img, 20, 80, frameStr)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addLabel(img *image.RGBA, x, y int, label string) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// Simple end-to-end latency test
func runEndToEndTest() {
	fmt.Println("\n=== End-to-End Latency Test ===")
	fmt.Println("This test measures the time from file write to stream receipt\n")

	// Create initial image
	generator := &ImageGenerator{}
	imgData, _ := generator.GenerateImage(time.Now())
	os.WriteFile(testImagePath, imgData, 0644)

	// Wait for server to pick it up
	fmt.Println("Waiting for server to initialize...")
	time.Sleep(200 * time.Millisecond)

	// Connect to stream
	fmt.Println("Connecting to video stream...")
	resp, err := http.Get(serverURL)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v\nMake sure the server is running on port 8080", err)
	}
	defer resp.Body.Close()

	mediaType, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(mediaType, "multipart/") {
		log.Fatalf("Expected multipart stream, got: %s", mediaType)
	}

	reader := multipart.NewReader(resp.Body, params["boundary"])

	// Consume first frame
	fmt.Println("Consuming initial frame...")
	reader.NextPart()

	fmt.Println("\nWrite Time              | Receive Time            | Latency (ms)")
	fmt.Println("------------------------|-------------------------|-------------")

	var latencies []time.Duration

	// Do 10 test writes and measure latency
	for i := 0; i < 10; i++ {
		writeTime := time.Now()
		imgData, _ := generator.GenerateImage(writeTime)
		os.WriteFile(testImagePath, imgData, 0644)

		// Wait for new frame in stream
		part, err := reader.NextPart()
		if err != nil {
			log.Printf("Error reading part: %v", err)
			continue
		}

		io.ReadAll(part) // Consume the data
		receiveTime := time.Now()

		latency := receiveTime.Sub(writeTime)
		latencies = append(latencies, latency)

		fmt.Printf("%s | %s | %8.2f ms\n",
			writeTime.Format("15:04:05.000000"),
			receiveTime.Format("15:04:05.000000"),
			float64(latency.Microseconds())/1000.0)

		time.Sleep(150 * time.Millisecond) // Wait between tests
	}

	// Print statistics
	if len(latencies) > 0 {
		var total time.Duration
		min := latencies[0]
		max := latencies[0]

		for _, lat := range latencies {
			total += lat
			if lat < min {
				min = lat
			}
			if lat > max {
				max = lat
			}
		}

		avg := total / time.Duration(len(latencies))

		fmt.Println("\n=== Latency Statistics ===")
		fmt.Printf("Min:     %8.2f ms\n", float64(min.Microseconds())/1000.0)
		fmt.Printf("Max:     %8.2f ms\n", float64(max.Microseconds())/1000.0)
		fmt.Printf("Average: %8.2f ms\n", float64(avg.Microseconds())/1000.0)
		fmt.Printf("Samples: %d\n", len(latencies))
	}
}

func main() {
	runEndToEndTest()
}
