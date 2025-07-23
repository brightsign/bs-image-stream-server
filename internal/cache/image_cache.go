package cache

import (
	"fmt"
	"sync"
	"time"
)

type ImageCache struct {
	mu       sync.RWMutex
	data     []byte
	modTime  time.Time
	etag     string
	fileSize int64
}

func NewImageCache() *ImageCache {
	return &ImageCache{}
}

func (c *ImageCache) Update(data []byte, modTime time.Time, fileSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data = make([]byte, len(data))
	copy(c.data, data)
	c.modTime = modTime
	c.fileSize = fileSize
	c.etag = fmt.Sprintf("\"%d-%d\"", modTime.Unix(), fileSize)
}

func (c *ImageCache) Get() ([]byte, string, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.data == nil {
		return nil, "", time.Time{}, false
	}
	
	dataCopy := make([]byte, len(c.data))
	copy(dataCopy, c.data)
	
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