package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
)

type Server struct {
	port       int
	cache      *cache.ImageCache
	httpServer *http.Server
}

func NewServer(port int, cache *cache.ImageCache) *Server {
	return &Server{
		port:  port,
		cache: cache,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/image", s.handleImage)
	mux.HandleFunc("/video", s.handleVideo)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/images/brightsign-logo.svg", s.handleLogo)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.HandleFunc("/", s.handleIndex)

	s.httpServer = &http.Server{
		Addr:        fmt.Sprintf(":%d", s.port),
		Handler:     mux,
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
