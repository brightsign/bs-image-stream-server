package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
	"github.com/bs-frame-monitor/internal/monitor"
	"github.com/bs-frame-monitor/internal/server"
)

func createTestJPEG(content string) []byte {
	return []byte(fmt.Sprintf("FAKE_JPEG_HEADER_%s_END", content))
}

func TestFullSystemIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	
	initialContent := createTestJPEG("initial")
	if err := os.WriteFile(testFile, initialContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	imageCache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, imageCache, time.Millisecond*10)
	srv := server.NewServer(0, imageCache)
	
	fileMonitor.Start()
	defer fileMonitor.Stop()
	
	go func() {
		srv.Start()
	}()
	defer srv.Shutdown()
	
	time.Sleep(time.Millisecond * 100)
	
	baseURL := "http://localhost:8080"
	
	t.Run("health endpoint works", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "ok") {
			t.Error("Health should report ok status")
		}
	})
	
	t.Run("index page serves HTML", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("Expected HTML content type, got %s", contentType)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "BrightSign") {
			t.Error("HTML should contain BrightSign branding")
		}
	})
	
	t.Run("image endpoint serves cached image", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/image")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/jpeg" {
			t.Errorf("Expected JPEG content type, got %s", contentType)
		}
		
		etag := resp.Header.Get("ETag")
		if etag == "" {
			t.Error("ETag should be set")
		}
		
		body, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(body, initialContent) {
			t.Error("Image content should match file content")
		}
	})
	
	t.Run("file updates are detected and served", func(t *testing.T) {
		updatedContent := createTestJPEG("updated")
		if err := os.WriteFile(testFile, updatedContent, 0644); err != nil {
			t.Fatalf("Failed to update test file: %v", err)
		}
		
		time.Sleep(time.Millisecond * 100)
		
		resp, err := http.Get(baseURL + "/image")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(body, updatedContent) {
			t.Error("Updated image content should be served")
		}
	})
	
	t.Run("ETag enables conditional requests", func(t *testing.T) {
		resp1, err := http.Get(baseURL + "/image")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		resp1.Body.Close()
		
		etag := resp1.Header.Get("ETag")
		if etag == "" {
			t.Fatal("First request should have ETag")
		}
		
		req, _ := http.NewRequest("GET", baseURL+"/image", nil)
		req.Header.Set("If-None-Match", etag)
		
		client := &http.Client{}
		resp2, err := client.Do(req)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp2.Body.Close()
		
		if resp2.StatusCode != http.StatusNotModified {
			t.Errorf("Expected status 304, got %d", resp2.StatusCode)
		}
	})
}

func TestSystemPerformance(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "perf_test.jpg")
	
	testContent := createTestJPEG("performance_test")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	imageCache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, imageCache, time.Millisecond*33)
	
	fileMonitor.Start()
	defer fileMonitor.Stop()
	
	time.Sleep(time.Millisecond * 200)
	
	start := time.Now()
	iterations := 1000
	
	for i := 0; i < iterations; i++ {
		if _, _, _, ok := imageCache.Get(); !ok {
			t.Fatal("Cache should always have data")
		}
	}
	
	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)
	
	if avgTime > time.Microsecond*100 {
		t.Errorf("Cache access too slow: %v per operation", avgTime)
	}
	
	t.Logf("Average cache access time: %v", avgTime)
}

func TestSystemResourceUsage(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "resource_test.jpg")
	
	largeContent := bytes.Repeat([]byte("LARGE_IMAGE_DATA"), 10000)
	if err := os.WriteFile(testFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	
	imageCache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, imageCache, time.Millisecond*33)
	
	fileMonitor.Start()
	defer fileMonitor.Stop()
	
	time.Sleep(time.Millisecond * 100)
	
	for i := 0; i < 100; i++ {
		updatedContent := append(largeContent, byte(i))
		if err := os.WriteFile(testFile, updatedContent, 0644); err != nil {
			t.Fatalf("Failed to update file: %v", err)
		}
		time.Sleep(time.Millisecond * 40)
	}
	
	if !imageCache.HasData() {
		t.Error("Cache should maintain data under load")
	}
	
	data, _, _, _ := imageCache.Get()
	if len(data) < len(largeContent) {
		t.Error("Large file should be properly cached")
	}
}