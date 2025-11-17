package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
)

type Server struct {
	port       int
	cache      *cache.ImageCache
	httpServer *http.Server
	debug      bool
}

func NewServer(port int, cache *cache.ImageCache, debug bool) *Server {
	return &Server{
		port:  port,
		cache: cache,
		debug: debug,
	}
}

// loggingMiddleware logs all HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("[SERVER] → %s %s from %s (User-Agent: %s)",
			r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Create a custom ResponseWriter to capture the status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// For streaming endpoints, log completion regardless of duration
		// For other endpoints, duration will be short
		log.Printf("[SERVER] ← %s %s completed with status %d in %v",
			r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/image", s.handleImage)
	mux.HandleFunc("/video", s.handleVideo)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/images/brightsign-logo.svg", s.handleLogo)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.HandleFunc("/", s.handleIndex)

	// Wrap with logging middleware
	handler := s.loggingMiddleware(mux)

	s.httpServer = &http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     handler,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout must be 0 for long-lived streaming connections
		// The multipart video stream writes continuously
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
}
