package monitor

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
	"github.com/fsnotify/fsnotify"
)

type FileMonitor struct {
	filePath string
	cache    *cache.ImageCache
	watcher  *fsnotify.Watcher
	stopCh   chan struct{}
}

func NewFileMonitor(filePath string, cache *cache.ImageCache, _ time.Duration) *FileMonitor {
	return &FileMonitor{
		filePath: filePath,
		cache:    cache,
		stopCh:   make(chan struct{}),
	}
}

func (fm *FileMonitor) Start() {
	var err error
	fm.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}

	// Load initial image if it exists
	fm.readAndCacheImage()

	// Watch the directory, not the file (file may be recreated)
	dir := filepath.Dir(fm.filePath)
	if err := fm.watcher.Add(dir); err != nil {
		log.Fatalf("Failed to watch directory %s: %v", dir, err)
	}

	log.Printf("Watching for changes to %s", fm.filePath)

	go func() {
		for {
			select {
			case event, ok := <-fm.watcher.Events:
				if !ok {
					return
				}

				// Only process events for our specific file
				if event.Name != fm.filePath {
					continue
				}

				// React to file writes and creates
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// Small delay to ensure write is complete
					time.Sleep(5 * time.Millisecond)
					fm.readAndCacheImage()
				}

			case err, ok := <-fm.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watcher error: %v", err)

			case <-fm.stopCh:
				return
			}
		}
	}()
}

func (fm *FileMonitor) Stop() {
	if fm.watcher != nil {
		fm.watcher.Close()
	}
	close(fm.stopCh)
}

func (fm *FileMonitor) readAndCacheImage() {
	stat, err := os.Stat(fm.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error checking file %s: %v", fm.filePath, err)
		}
		return
	}

	file, err := os.Open(fm.filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", fm.filePath, err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file %s: %v", fm.filePath, err)
		return
	}

	modTime := stat.ModTime()
	fm.cache.Update(data, modTime, stat.Size())
}
