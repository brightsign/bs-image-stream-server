package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
)

func createTempFile(t *testing.T, content []byte) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.jpg")
	
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	return filePath
}

func TestNewFileMonitor(t *testing.T) {
	cache := cache.NewImageCache()
	filePath := "/tmp/test.jpg"
	interval := time.Millisecond * 50
	
	monitor := NewFileMonitor(filePath, cache, interval)
	
	if monitor == nil {
		t.Fatal("NewFileMonitor returned nil")
	}
	
	if monitor.filePath != filePath {
		t.Errorf("Expected filePath %s, got %s", filePath, monitor.filePath)
	}
	
	if monitor.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, monitor.interval)
	}
}

func TestFileMonitorDetectsNewFile(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("test image content")
	filePath := createTempFile(t, testData)
	
	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()
	
	time.Sleep(time.Millisecond * 50)
	
	if !cache.HasData() {
		t.Error("Cache should have data after file monitor starts")
	}
	
	data, _, _, ok := cache.Get()
	if !ok {
		t.Fatal("Cache Get() should return true")
	}
	
	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got %s", string(testData), string(data))
	}
}

func TestFileMonitorDetectsFileChanges(t *testing.T) {
	cache := cache.NewImageCache()
	initialData := []byte("initial content")
	filePath := createTempFile(t, initialData)
	
	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()
	
	time.Sleep(time.Millisecond * 50)
	
	if !cache.HasData() {
		t.Error("Cache should have initial data")
	}
	
	updatedData := []byte("updated content")
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	
	time.Sleep(time.Millisecond * 50)
	
	data, _, _, ok := cache.Get()
	if !ok {
		t.Fatal("Cache Get() should return true after update")
	}
	
	if string(data) != string(updatedData) {
		t.Errorf("Expected updated data %s, got %s", string(updatedData), string(data))
	}
}

func TestFileMonitorHandlesMissingFile(t *testing.T) {
	cache := cache.NewImageCache()
	filePath := "/tmp/nonexistent.jpg"
	
	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()
	
	time.Sleep(time.Millisecond * 50)
	
	if cache.HasData() {
		t.Error("Cache should not have data for missing file")
	}
}

func TestFileMonitorStopPreventsUpdates(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("test content")
	filePath := createTempFile(t, testData)
	
	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	
	time.Sleep(time.Millisecond * 50)
	
	if !cache.HasData() {
		t.Error("Cache should have data after start")
	}
	
	monitor.Stop()
	
	updatedData := []byte("should not be cached")
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	
	time.Sleep(time.Millisecond * 50)
	
	data, _, _, _ := cache.Get()
	if string(data) == string(updatedData) {
		t.Error("Cache should not have updated after stop")
	}
}

func TestFileMonitorIgnoresUnchangedFile(t *testing.T) {
	cache := cache.NewImageCache()
	testData := []byte("unchanging content")
	filePath := createTempFile(t, testData)
	
	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()
	
	time.Sleep(time.Millisecond * 50)
	
	if !cache.HasData() {
		t.Error("Cache should have initial data")
	}
	
	_, _, firstModTime, _ := cache.Get()
	
	time.Sleep(time.Millisecond * 100)
	
	_, _, secondModTime, _ := cache.Get()
	
	if !firstModTime.Equal(secondModTime) {
		t.Error("Modification time should not change for unchanged file")
	}
}