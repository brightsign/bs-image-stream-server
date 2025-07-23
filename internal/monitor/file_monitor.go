package monitor

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
)

type FileMonitor struct {
	filePath    string
	cache       *cache.ImageCache
	ticker      *time.Ticker
	stopCh      chan struct{}
	interval    time.Duration
	lastModTime time.Time
}

func NewFileMonitor(filePath string, cache *cache.ImageCache, interval time.Duration) *FileMonitor {
	return &FileMonitor{
		filePath: filePath,
		cache:    cache,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (fm *FileMonitor) Start() {
	fm.ticker = time.NewTicker(fm.interval)
	
	fm.checkAndUpdateImage()
	
	go func() {
		for {
			select {
			case <-fm.ticker.C:
				fm.checkAndUpdateImage()
			case <-fm.stopCh:
				return
			}
		}
	}()
}

func (fm *FileMonitor) Stop() {
	if fm.ticker != nil {
		fm.ticker.Stop()
	}
	close(fm.stopCh)
}

func (fm *FileMonitor) checkAndUpdateImage() {
	stat, err := os.Stat(fm.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error checking file %s: %v", fm.filePath, err)
		}
		return
	}

	modTime := stat.ModTime()
	if modTime.Equal(fm.lastModTime) {
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

	fm.cache.Update(data, modTime, stat.Size())
	fm.lastModTime = modTime
}