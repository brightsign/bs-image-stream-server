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

// createValidJPEG creates minimal valid JPEG data for testing
// JPEG format: SOI (FF D8) + data + EOI (FF D9)
func createValidJPEG(payload string) []byte {
	data := []byte{0xFF, 0xD8} // SOI marker
	data = append(data, []byte(payload)...)
	data = append(data, 0xFF, 0xD9) // EOI marker
	return data
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
}

func TestFileMonitorDetectsNewFile(t *testing.T) {
	cache := cache.NewImageCache()
	testData := createValidJPEG("test image content")
	filePath := createTempFile(t, testData)

	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()

	// Give watcher time to load initial file
	time.Sleep(time.Millisecond * 100)

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
	initialData := createValidJPEG("initial content")
	filePath := createTempFile(t, initialData)

	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()

	// Give watcher time to load initial file
	time.Sleep(time.Millisecond * 100)

	if !cache.HasData() {
		t.Error("Cache should have initial data")
	}

	updatedData := createValidJPEG("updated content")
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Give watcher time to detect and process the change
	time.Sleep(time.Millisecond * 100)

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
	testData := createValidJPEG("test content")
	filePath := createTempFile(t, testData)

	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()

	// Give watcher time to load initial file
	time.Sleep(time.Millisecond * 100)

	if !cache.HasData() {
		t.Error("Cache should have data after start")
	}

	monitor.Stop()

	updatedData := createValidJPEG("should not be cached")
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Give time for any stale events to process
	time.Sleep(time.Millisecond * 100)

	data, _, _, _ := cache.Get()
	if string(data) == string(updatedData) {
		t.Error("Cache should not have updated after stop")
	}
}

func TestFileMonitorIgnoresUnchangedFile(t *testing.T) {
	cache := cache.NewImageCache()
	testData := createValidJPEG("unchanging content")
	filePath := createTempFile(t, testData)

	monitor := NewFileMonitor(filePath, cache, time.Millisecond*10)
	monitor.Start()
	defer monitor.Stop()

	// Give watcher time to load initial file
	time.Sleep(time.Millisecond * 100)

	if !cache.HasData() {
		t.Error("Cache should have initial data")
	}

	_, _, firstModTime, _ := cache.Get()

	// Wait without file changes
	time.Sleep(time.Millisecond * 200)

	_, _, secondModTime, _ := cache.Get()

	if !firstModTime.Equal(secondModTime) {
		t.Error("Modification time should not change for unchanged file")
	}
}
