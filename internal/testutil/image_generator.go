package testutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

var jpegHeader = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
}

var jpegFooter = []byte{0xFF, 0xD9}

func GenerateTestJPEG(width, height int, content string) []byte {
	var buf bytes.Buffer

	buf.Write(jpegHeader)

	buf.WriteString(fmt.Sprintf("TEST_IMAGE_%s_", content))

	buf.Write([]byte{0xFF, 0xC0, 0x00, 0x11, 0x08})
	binary.Write(&buf, binary.BigEndian, uint16(height))
	binary.Write(&buf, binary.BigEndian, uint16(width))
	buf.Write([]byte{0x03, 0x01, 0x22, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01})

	dataSize := width * height * 3 / 10
	for i := 0; i < dataSize; i++ {
		buf.WriteByte(byte(rand.Intn(256)))
	}

	buf.Write(jpegFooter)

	return buf.Bytes()
}

func GenerateTestJPEGWithTimestamp() []byte {
	timestamp := time.Now().Format("20060102_150405")
	return GenerateTestJPEG(640, 480, timestamp)
}

func GenerateRandomJPEG(minSize, maxSize int) []byte {
	size := minSize + rand.Intn(maxSize-minSize)
	content := fmt.Sprintf("RANDOM_%d", rand.Intn(10000))

	var buf bytes.Buffer
	buf.Write(jpegHeader)
	buf.WriteString(content)

	for buf.Len() < size-len(jpegFooter) {
		buf.WriteByte(byte(rand.Intn(256)))
	}

	buf.Write(jpegFooter)
	return buf.Bytes()
}

func GenerateLargeTestJPEG(sizeMB int) []byte {
	targetSize := sizeMB * 1024 * 1024
	content := fmt.Sprintf("LARGE_%dMB", sizeMB)

	var buf bytes.Buffer
	buf.Write(jpegHeader)
	buf.WriteString(content)

	chunkSize := 1024
	chunk := make([]byte, chunkSize)
	for i := range chunk {
		chunk[i] = byte(rand.Intn(256))
	}

	for buf.Len() < targetSize-len(jpegFooter) {
		remaining := targetSize - len(jpegFooter) - buf.Len()
		if remaining < chunkSize {
			buf.Write(chunk[:remaining])
		} else {
			buf.Write(chunk)
		}
	}

	buf.Write(jpegFooter)
	return buf.Bytes()
}

func ValidateJPEGFormat(data []byte) bool {
	if len(data) < len(jpegHeader)+len(jpegFooter) {
		return false
	}

	if !bytes.HasPrefix(data, jpegHeader[:4]) {
		return false
	}

	if !bytes.HasSuffix(data, jpegFooter) {
		return false
	}

	return true
}
