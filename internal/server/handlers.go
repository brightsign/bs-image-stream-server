package server

import (
	"embed"
	"fmt"
	"log"
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
	if s.debug {
		log.Printf("[HANDLER] /image request from %s, If-None-Match=%s",
			r.RemoteAddr, r.Header.Get("If-None-Match"))
	}

	data, etag, modTime, ok := s.cache.Get()
	if !ok {
		if s.debug {
			log.Printf("[HANDLER] /image: No image data available, returning 404")
		}
		http.Error(w, "Image not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "no-cache")

	clientETag := r.Header.Get("If-None-Match")
	if clientETag == etag {
		if s.debug {
			log.Printf("[HANDLER] /image: Client ETag matches, returning 304 Not Modified (etag=%s)", etag)
		}
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if s.debug {
		log.Printf("[HANDLER] /image: Sending image data (%d bytes, etag=%s, client_etag=%s)",
			len(data), etag, clientETag)
	}

	w.Write(data)
}

func (s *Server) handleVideo(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	log.Printf("[HANDLER] handleVideo() called - format=%q, client=%s", format, r.RemoteAddr)

	switch format {
	case "mjpeg":
		log.Printf("[HANDLER] Routing to MJPEG stream handler")
		s.handleMJPEGStream(w, r)
	default:
		// Default to multipart stream for browser compatibility
		log.Printf("[HANDLER] Routing to multipart stream handler")
		s.handleMultipartStream(w, r)
	}
}

func (s *Server) handleMultipartStream(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] ▶ Video stream handler called for client %s", r.RemoteAddr)

	// Set multipart/x-mixed-replace header for streaming
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	log.Printf("[HANDLER] ▶ Video stream STARTED for client %s (headers sent, beginning frame loop)", r.RemoteAddr)

	// Stream images at 30 FPS
	ticker := time.NewTicker(time.Millisecond * 33)
	defer ticker.Stop()

	frameCount := 0
	startTime := time.Now()
	var lastETag string

	for {
		select {
		case <-r.Context().Done():
			// Log why the stream ended
			duration := time.Since(startTime)
			log.Printf("[HANDLER] Video stream ended for client %s | Duration: %v | Frames sent: %d | Reason: %v",
				r.RemoteAddr, duration, frameCount, r.Context().Err())
			return
		case <-ticker.C:
			data, etag, _, ok := s.cache.Get()
			if !ok {
				if s.debug {
					log.Printf("[HANDLER] Video stream: No image data available for frame %d", frameCount)
				}
				// No image data available yet - wait for next tick
				continue
			}

			// Log frame info in debug mode (every 30 frames = ~1 second)
			// or when ETag changes (new frame detected)
			if s.debug {
				if etag != lastETag {
					log.Printf("[HANDLER] Video stream [%s]: Frame %d - NEW FRAME (etag=%s, size=%d bytes, prev_etag=%s)",
						r.RemoteAddr, frameCount, etag, len(data), lastETag)
					lastETag = etag
				} else if frameCount%30 == 0 {
					log.Printf("[HANDLER] Video stream [%s]: Frame %d - REUSING (etag=%s, size=%d bytes)",
						r.RemoteAddr, frameCount, etag, len(data))
				}
			}

			// Log heartbeat every 300 frames (~10 seconds) even without debug
			if frameCount > 0 && frameCount%300 == 0 {
				duration := time.Since(startTime)
				fps := float64(frameCount) / duration.Seconds()
				log.Printf("[HANDLER] Video stream [%s]: Heartbeat - %d frames sent, %.1f FPS, duration=%v",
					r.RemoteAddr, frameCount, fps, duration)
			}

			// Write multipart boundary and headers
			_, err := w.Write([]byte("--frame\r\n"))
			if err != nil {
				log.Printf("[HANDLER] Video stream boundary write error for client %s (frame %d): %v", r.RemoteAddr, frameCount, err)
				return
			}

			_, err = w.Write([]byte(fmt.Sprintf("Content-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(data))))
			if err != nil {
				log.Printf("[HANDLER] Video stream header write error for client %s (frame %d): %v", r.RemoteAddr, frameCount, err)
				return
			}

			// Write image data
			_, err = w.Write(data)
			if err != nil {
				log.Printf("[HANDLER] Video stream data write error for client %s (frame %d): %v", r.RemoteAddr, frameCount, err)
				return
			}

			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				log.Printf("[HANDLER] Video stream trailing CRLF write error for client %s (frame %d): %v", r.RemoteAddr, frameCount, err)
				return
			}

			// Flush to send immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			} else if s.debug && frameCount == 0 {
				log.Printf("[HANDLER] Video stream [%s]: WARNING - ResponseWriter does not support flushing!", r.RemoteAddr)
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

	updateCount, lastUpdate, hasData := s.cache.GetStats()

	status := "ok"
	if !hasData {
		status = "no_image"
	}

	// Calculate time since last update
	timeSinceUpdate := time.Since(lastUpdate)
	staleness := "fresh"
	if timeSinceUpdate > 5*time.Second {
		staleness = "stale"
	}

	response := fmt.Sprintf(`{"status":"%s","timestamp":"%s","frames_cached":%d,"last_update":"%s","time_since_update_ms":%d,"staleness":"%s"}`,
		status,
		time.Now().UTC().Format(time.RFC3339),
		updateCount,
		lastUpdate.UTC().Format(time.RFC3339),
		timeSinceUpdate.Milliseconds(),
		staleness,
	)

	w.Write([]byte(response))
}

func (s *Server) handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write(brightSignLogo)
}
