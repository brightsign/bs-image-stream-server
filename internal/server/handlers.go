package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"
)

//go:embed static
var staticFiles embed.FS

//go:embed static/brightsign-logo.svg
var brightSignLogo []byte

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(data)
}

func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	data, etag, modTime, ok := s.cache.Get()
	if !ok {
		http.Error(w, "Image not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "no-cache")

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Write(data)
}

func (s *Server) handleVideo(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	
	switch format {
	case "mjpeg":
		s.handleMJPEGStream(w, r)
	default:
		// Default to multipart stream for browser compatibility
		s.handleMultipartStream(w, r)
	}
}

func (s *Server) handleMultipartStream(w http.ResponseWriter, r *http.Request) {
	// Set multipart/x-mixed-replace header for streaming
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Stream images at 30 FPS
	ticker := time.NewTicker(time.Millisecond * 33)
	defer ticker.Stop()

	frameCount := 0

	for {
		select {
		case <-r.Context().Done():
			// Log why the stream ended
			if r.Context().Err() != nil {
				// Client disconnected or request cancelled - this is normal
			}
			return
		case <-ticker.C:
			data, _, _, ok := s.cache.Get()
			if !ok {
				// No image data available yet - wait for next tick
				continue
			}

			// Write multipart boundary and headers
			_, err := w.Write([]byte("--frame\r\n"))
			if err != nil {
				// Connection closed by client
				return
			}

			_, err = w.Write([]byte(fmt.Sprintf("Content-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(data))))
			if err != nil {
				return
			}

			// Write image data
			_, err = w.Write(data)
			if err != nil {
				return
			}

			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}

			// Flush to send immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			frameCount++
		}
	}
}

func (s *Server) handleMJPEGStream(w http.ResponseWriter, r *http.Request) {
	// Use the same multipart format as the main stream for consistency
	// This provides better ffmpeg compatibility
	s.handleMultipartStream(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	if !s.cache.HasData() {
		status = "no_image"
	}

	w.Write([]byte(`{"status":"` + status + `","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

func (s *Server) handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write(brightSignLogo)
}
