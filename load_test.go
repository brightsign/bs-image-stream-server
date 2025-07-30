package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
	"github.com/bs-frame-monitor/internal/monitor"
)

func BenchmarkImageCacheGet(b *testing.B) {
	cache := cache.NewImageCache()
	testData := bytes.Repeat([]byte("test"), 1000)
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get()
		}
	})
}

func BenchmarkImageCacheUpdate(b *testing.B) {
	cache := cache.NewImageCache()
	testData := bytes.Repeat([]byte("test"), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modTime := time.Now().Add(time.Duration(i) * time.Nanosecond)
		cache.Update(testData, modTime, int64(len(testData)))
	}
}

func BenchmarkFileMonitorWithFrequentUpdates(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.jpg")

	testData := bytes.Repeat([]byte("benchmark"), 1000)
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	cache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, cache, time.Millisecond*33)

	fileMonitor.Start()
	defer fileMonitor.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		updatedData := append(testData, byte(i%256))
		if err := os.WriteFile(testFile, updatedData, 0644); err != nil {
			b.Fatalf("Failed to update file: %v", err)
		}
		time.Sleep(time.Millisecond * 35)
	}
}

func BenchmarkHTTPImageHandler(b *testing.B) {
	cache := cache.NewImageCache()
	testData := bytes.Repeat([]byte("image_data"), 10000)
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	mux := http.NewServeMux()
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		data, etag, lastMod, ok := cache.Get()
		if !ok {
			http.Error(w, "Image not available", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastMod.UTC().Format(http.TimeFormat))
		w.Write(data)
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/image", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
		}
	})
}

func TestConcurrentImageRequests(t *testing.T) {
	cache := cache.NewImageCache()
	testData := bytes.Repeat([]byte("concurrent_test"), 5000)
	modTime := time.Now()

	cache.Update(testData, modTime, int64(len(testData)))

	mux := http.NewServeMux()
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		data, etag, lastMod, ok := cache.Get()
		if !ok {
			http.Error(w, "Image not available", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastMod.UTC().Format(http.TimeFormat))
		w.Write(data)
	})
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	concurrency := 50
	requests := 100

	var wg sync.WaitGroup
	errors := make(chan error, concurrency*requests)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client := &http.Client{Timeout: time.Second * 5}

			for j := 0; j < requests; j++ {
				resp, err := client.Get(testServer.URL + "/image")
				if err != nil {
					errors <- fmt.Errorf("request failed: %v", err)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				}

				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)
	totalRequests := concurrency * requests
	rps := float64(totalRequests) / duration.Seconds()

	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Had %d errors out of %d requests", errorCount, totalRequests)
	}

	t.Logf("Completed %d concurrent requests in %v (%.2f req/sec)", totalRequests, duration, rps)

	if rps < 100 {
		t.Errorf("Performance too low: %.2f req/sec (expected > 100)", rps)
	}
}

func TestMemoryUsageUnderLoad(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "memory_test.jpg")

	largeImageData := bytes.Repeat([]byte("LARGE_IMAGE_PIXEL_DATA"), 50000)
	if err := os.WriteFile(testFile, largeImageData, 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	cache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, cache, time.Millisecond*33)

	fileMonitor.Start()
	defer fileMonitor.Stop()

	time.Sleep(time.Millisecond * 100)

	var wg sync.WaitGroup
	concurrency := 20
	duration := time.Second * 2

	stop := make(chan struct{})
	time.AfterFunc(duration, func() { close(stop) })

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			requestCount := 0
			for {
				select {
				case <-stop:
					t.Logf("Goroutine completed %d requests", requestCount)
					return
				default:
					if data, _, _, ok := cache.Get(); !ok || len(data) != len(largeImageData) {
						t.Errorf("Cache data inconsistent")
						return
					}
					requestCount++
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()

	if !cache.HasData() {
		t.Error("Cache should still have data after load test")
	}
}

func TestSystemStabilityUnder30FPS(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fps_test.jpg")

	baseData := bytes.Repeat([]byte("30FPS_TEST"), 1000)
	if err := os.WriteFile(testFile, baseData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(testFile, cache, time.Millisecond*33)

	fileMonitor.Start()
	defer fileMonitor.Stop()

	updates := 90
	interval := time.Millisecond * 33

	start := time.Now()

	for i := 0; i < updates; i++ {
		updatedData := append(baseData, []byte(fmt.Sprintf("_FRAME_%d", i))...)
		if err := os.WriteFile(testFile, updatedData, 0644); err != nil {
			t.Fatalf("Failed to update file at frame %d: %v", i, err)
		}
		time.Sleep(interval)
	}

	duration := time.Since(start)
	expectedDuration := time.Duration(updates) * interval

	if duration > expectedDuration*2 {
		t.Errorf("Test took too long: %v (expected ~%v)", duration, expectedDuration)
	}

	actualFPS := float64(updates) / duration.Seconds()
	expectedFPS := 30.0

	if actualFPS < expectedFPS*0.8 {
		t.Errorf("FPS too low: %.2f (expected ~%.2f)", actualFPS, expectedFPS)
	}

	t.Logf("Achieved %.2f FPS over %v", actualFPS, duration)
}
