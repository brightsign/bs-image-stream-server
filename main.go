package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bs-frame-monitor/internal/cache"
	"github.com/bs-frame-monitor/internal/monitor"
	"github.com/bs-frame-monitor/internal/server"
)

func main() {
	var (
		port     = flag.Int("port", 8080, "HTTP server port")
		filePath = flag.String("file", "/tmp/output.jpg", "Path to image file to monitor")
		debug    = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	imageCache := cache.NewImageCache()
	fileMonitor := monitor.NewFileMonitor(*filePath, imageCache, time.Millisecond*33)

	fileMonitor.Start()
	defer fileMonitor.Stop()

	srv := server.NewServer(*port, imageCache)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		srv.Shutdown()
		os.Exit(0)
	}()

	log.Printf("Starting BS Frame Monitor on port %d, monitoring %s", *port, *filePath)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
