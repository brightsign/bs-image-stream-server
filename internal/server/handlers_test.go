package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
)

func TestHandleIndex(t *testing.T) {
	cache := cache.NewImageCache()
	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "BrightSign Image Stream Server") {
		t.Error("Response should contain page title")
	}

	if !strings.Contains(body, "setInterval") {
		t.Error("Response should contain JavaScript for image refresh")
	}
}

func TestHandleIndexNotFound(t *testing.T) {
	cache := cache.NewImageCache()
	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleImageNoData(t *testing.T) {
	cache := cache.NewImageCache()
	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/image", nil)
	w := httptest.NewRecorder()

	server.handleImage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleImageWithData(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("fake jpeg data")
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/image", nil)
	w := httptest.NewRecorder()

	server.handleImage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "image/jpeg" {
		t.Errorf("Expected JPEG content type, got %s", contentType)
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("ETag header should be set")
	}

	lastModified := w.Header().Get("Last-Modified")
	if lastModified == "" {
		t.Error("Last-Modified header should be set")
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control: no-cache, got %s", cacheControl)
	}

	if w.Body.String() != string(testData) {
		t.Error("Response body should match cached data")
	}
}

func TestHandleImageNotModified(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("fake jpeg data")
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	server := NewServer(8080, cache)

	data, etag, _, _ := cache.Get()
	if len(data) == 0 {
		t.Fatal("Cache should have data")
	}

	req := httptest.NewRequest("GET", "/image", nil)
	req.Header.Set("If-None-Match", etag)
	w := httptest.NewRecorder()

	server.handleImage(w, req)

	if w.Code != http.StatusNotModified {
		t.Errorf("Expected status 304, got %d", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Error("304 response should have empty body")
	}
}

func TestHandleHealth(t *testing.T) {
	cache := cache.NewImageCache()
	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"status":"no_image"`) {
		t.Error("Health should report no_image when cache is empty")
	}

	if !strings.Contains(body, `"timestamp"`) {
		t.Error("Health response should include timestamp")
	}
}

func TestHandleHealthWithImage(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("test image")
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	server := NewServer(8080, cache)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Error("Health should report ok when cache has data")
	}
}

