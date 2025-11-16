package monitor

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
	"github.com/fsnotify/fsnotify"
)

type FileMonitor struct {
	filePath     string
	cache        *cache.ImageCache
	watcher      *fsnotify.Watcher
	stopCh       chan struct{}
	debug        bool
	lastModTime  time.Time
	lastFileSize int64
}

func NewFileMonitor(filePath string, cache *cache.ImageCache, _ time.Duration, debug bool) *FileMonitor {
	return &FileMonitor{
		filePath: filePath,
		cache:    cache,
		stopCh:   make(chan struct{}),
		debug:    debug,
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
					if fm.debug {
						log.Printf("[MONITOR] Ignoring event for %s (watching %s)", event.Name, fm.filePath)
					}
					continue
				}

				if fm.debug {
					log.Printf("[MONITOR] File event detected: %s on %s", event.Op, event.Name)
				}

				// React to file writes and creates
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if fm.debug {
						log.Printf("[MONITOR] Processing %s event for %s", event.Op, event.Name)
					}
					// Small delay to ensure write is complete
					time.Sleep(5 * time.Millisecond)
					fm.readAndCacheImage()
				}

			case err, ok := <-fm.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[MONITOR] File watcher error: %v", err)

			case <-fm.stopCh:
				if fm.debug {
					log.Printf("[MONITOR] Stop signal received, shutting down")
				}
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
	if fm.debug {
		log.Printf("[MONITOR] readAndCacheImage() called for %s", fm.filePath)
	}

	// Retry reading with exponential backoff to handle incomplete writes
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry: 5ms, 10ms, 20ms
			delay := time.Duration(5*(1<<uint(attempt-1))) * time.Millisecond
			if fm.debug {
				log.Printf("[MONITOR] Retry #%d after %v delay", attempt, delay)
			}
			time.Sleep(delay)
		}

		stat, err := os.Stat(fm.filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				if fm.debug {
					log.Printf("[MONITOR] Stat error on attempt #%d: %v", attempt, err)
				}
				lastErr = err
				continue
			}
			if fm.debug {
				log.Printf("[MONITOR] File does not exist: %s", fm.filePath)
			}
			return // File doesn't exist, don't spam logs
		}

		// Skip reading if file hasn't actually changed (deduplication)
		if stat.ModTime().Equal(fm.lastModTime) && stat.Size() == fm.lastFileSize {
			if fm.debug {
				log.Printf("[MONITOR] File unchanged (modTime=%s, size=%d), skipping read",
					stat.ModTime().Format(time.RFC3339Nano), stat.Size())
			}
			return
		}

		if fm.debug {
			log.Printf("[MONITOR] File stat: size=%d bytes, modTime=%s (prev: size=%d, modTime=%s)",
				stat.Size(), stat.ModTime().Format(time.RFC3339Nano),
				fm.lastFileSize, fm.lastModTime.Format(time.RFC3339Nano))
		}

		file, err := os.Open(fm.filePath)
		if err != nil {
			if fm.debug {
				log.Printf("[MONITOR] Open error on attempt #%d: %v", attempt, err)
			}
			lastErr = err
			continue
		}

		data, err := io.ReadAll(file)
		file.Close()

		if err != nil {
			if fm.debug {
				log.Printf("[MONITOR] Read error on attempt #%d: %v", attempt, err)
			}
			lastErr = err
			continue
		}

		if fm.debug {
			log.Printf("[MONITOR] Successfully read %d bytes from disk", len(data))
		}

		// Validate JPEG format - must have valid header and footer
		if !isValidJPEG(data) {
			if fm.debug {
				if len(data) < 4 {
					log.Printf("[MONITOR] Invalid JPEG on attempt #%d: size=%d bytes (too small, need at least 4 bytes)",
						attempt, len(data))
				} else {
					log.Printf("[MONITOR] Invalid JPEG on attempt #%d: size=%d bytes, header=%02x%02x, footer=%02x%02x",
						attempt, len(data),
						data[0], data[1],
						data[len(data)-2], data[len(data)-1])
				}
			}
			lastErr = fmt.Errorf("invalid JPEG data (size: %d bytes)", len(data))
			continue
		}

		// Successfully read and validated - update cache with new frame
		modTime := stat.ModTime()
		fileSize := stat.Size()
		if fm.debug {
			log.Printf("[MONITOR] Valid JPEG confirmed, updating cache with new frame")
		}
		fm.cache.Update(data, modTime, fileSize)

		// Update our tracking of last processed file state
		fm.lastModTime = modTime
		fm.lastFileSize = fileSize

		// Log only on retry success
		if attempt > 0 {
			log.Printf("[MONITOR] Successfully read image after %d retries", attempt)
		}
		return
	}

	// All retries failed - log error but keep serving last good frame
	if lastErr != nil {
		log.Printf("[MONITOR] Failed to read image after %d attempts (keeping last good frame): %v", maxRetries, lastErr)
		if fm.debug {
			log.Printf("[MONITOR] REUSING PREVIOUS FRAME due to read failure")
		}
	}
}

// isValidJPEG checks if the data is a valid JPEG image
func isValidJPEG(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check JPEG magic bytes: starts with FF D8, ends with FF D9
	hasValidHeader := data[0] == 0xFF && data[1] == 0xD8
	hasValidFooter := data[len(data)-2] == 0xFF && data[len(data)-1] == 0xD9

	return hasValidHeader && hasValidFooter
}
