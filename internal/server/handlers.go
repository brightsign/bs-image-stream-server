package server

import (
	"embed"
	"net/http"
	"time"
)

//go:embed static
var staticFiles embed.FS

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

func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
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
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := "ok"
	if !s.cache.HasData() {
		status = "no_image"
	}
	
	w.Write([]byte(`{"status":"` + status + `","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}