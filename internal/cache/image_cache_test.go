package cache

import (
	"bytes"
	"testing"
	"time"
)

func TestNewImageCache(t *testing.T) {
	cache := NewImageCache()
	if cache == nil {
		t.Fatal("NewImageCache returned nil")
	}

	if cache.HasData() {
		t.Error("New cache should not have data")
	}

	if etag := cache.GetETag(); etag != "" {
		t.Errorf("New cache should have empty ETag, got %s", etag)
	}
}

func TestImageCacheUpdate(t *testing.T) {
	cache := NewImageCache()
	testData := []byte("test image data")
	modTime := time.Now()
	fileSize := int64(len(testData))

	cache.Update(testData, modTime, fileSize)

	if !cache.HasData() {
		t.Error("Cache should have data after update")
	}

	data, etag, retrievedModTime, ok := cache.Get()
	if !ok {
		t.Fatal("Get() should return true after update")
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("Retrieved data doesn't match. Expected %v, got %v", testData, data)
	}

	if !retrievedModTime.Equal(modTime) {
		t.Errorf("Retrieved modTime doesn't match. Expected %v, got %v", modTime, retrievedModTime)
	}

	if len(etag) == 0 {
		t.Error("ETag should not be empty")
	}

	if cache.GetETag() != etag {
		t.Error("GetETag() should return same value as Get()")
	}
}

func TestImageCacheDataIsolation(t *testing.T) {
	cache := NewImageCache()
	testData := []byte("test image data")
	modTime := time.Now()
	fileSize := int64(len(testData))

	cache.Update(testData, modTime, fileSize)

	data1, _, _, _ := cache.Get()
	data2, _, _, _ := cache.Get()

	if &data1[0] == &data2[0] {
		t.Error("Get() should return independent copies of data")
	}

	data1[0] = 'X'
	data3, _, _, _ := cache.Get()

	if data3[0] == 'X' {
		t.Error("Modifying returned data should not affect cached data")
	}
}

func TestImageCacheMultipleUpdates(t *testing.T) {
	cache := NewImageCache()

	testData1 := []byte("first image")
	modTime1 := time.Now()
	cache.Update(testData1, modTime1, int64(len(testData1)))

	testData2 := []byte("second image data")
	modTime2 := modTime1.Add(time.Second)
	cache.Update(testData2, modTime2, int64(len(testData2)))

	data, _, modTime, ok := cache.Get()
	if !ok {
		t.Fatal("Get() should return true after updates")
	}

	if !bytes.Equal(data, testData2) {
		t.Errorf("Should retrieve latest data. Expected %v, got %v", testData2, data)
	}

	if !modTime.Equal(modTime2) {
		t.Errorf("Should retrieve latest modTime. Expected %v, got %v", modTime2, modTime)
	}
}

func TestImageCacheConcurrency(t *testing.T) {
	cache := NewImageCache()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			testData := []byte("concurrent test data")
			modTime := time.Now()
			cache.Update(testData, modTime, int64(len(testData)))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Get()
			cache.HasData()
			cache.GetETag()
		}
		done <- true
	}()

	<-done
	<-done

	if !cache.HasData() {
		t.Error("Cache should have data after concurrent operations")
	}
}
