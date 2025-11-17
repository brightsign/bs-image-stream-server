package cache

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type ImageCache struct {
	mu          sync.RWMutex
	data        []byte
	modTime     time.Time
	etag        string
	fileSize    int64
	updateCount uint64      // Total successful updates
	lastUpdate  time.Time   // Last successful update time
	debug       bool
}

func NewImageCache(debug bool) *ImageCache {
	return &ImageCache{
		debug: debug,
	}
}

func (c *ImageCache) Update(data []byte, modTime time.Time, fileSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldUpdateCount := c.updateCount
	oldModTime := c.modTime
	oldFileSize := c.fileSize

	// Always allocate new buffer to avoid aliasing issues
	c.data = make([]byte, len(data))
	copy(c.data, data)
	c.modTime = modTime
	c.fileSize = fileSize
	c.updateCount++
	// Include updateCount to ensure ETag is ALWAYS unique, even if modTime/fileSize don't change
	// This handles cases where filesystem fires multiple events for same file write
	c.etag = fmt.Sprintf("\"%d-%d-%d\"", modTime.UnixNano(), fileSize, c.updateCount)
	c.lastUpdate = time.Now()

	if c.debug {
		// Log when file metadata changed or every 30 updates (~1 second at 30 FPS)
		metadataChanged := oldUpdateCount == 0 ||
			!modTime.Equal(oldModTime) ||
			fileSize != oldFileSize

		if metadataChanged {
			log.Printf("[CACHE] UPDATE #%d: NEW FILE DATA - size=%d bytes, modTime=%s, etag=%s",
				c.updateCount, len(data), modTime.Format(time.RFC3339Nano), c.etag)
		} else if c.updateCount%30 == 0 {
			log.Printf("[CACHE] UPDATE #%d: Duplicate event - same modTime/size, etag=%s",
				c.updateCount, c.etag)
		}
	}
}

func (c *ImageCache) Get() ([]byte, string, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil {
		if c.debug {
			log.Printf("[CACHE] GET: No data available")
		}
		return nil, "", time.Time{}, false
	}

	dataCopy := make([]byte, len(c.data))
	copy(dataCopy, c.data)

	// Only log in debug mode if frame is stale
	if c.debug {
		age := time.Since(c.lastUpdate)
		if age > 1*time.Second {
			log.Printf("[CACHE] WARNING: Serving stale frame #%d (etag=%s, age=%v, last_update=%s)",
				c.updateCount, c.etag, age, c.lastUpdate.Format(time.RFC3339Nano))
		}
	}

	return dataCopy, c.etag, c.modTime, true
}

func (c *ImageCache) GetETag() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.etag
}

func (c *ImageCache) HasData() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data != nil
}

// GetStats returns cache statistics for monitoring
func (c *ImageCache) GetStats() (updateCount uint64, lastUpdate time.Time, hasData bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.updateCount, c.lastUpdate, c.data != nil
}
