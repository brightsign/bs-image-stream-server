# streamtest Makefile
# Test video streamer for bs-image-stream-server

BINARY_NAME=streamtest
BINARY_DIR=../bin

.PHONY: all
all: build

.PHONY: build
build:
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 go build -buildvcs=false -o $(BINARY_DIR)/$(BINARY_NAME) .

.PHONY: run
run: build
	@echo "Usage: $(BINARY_DIR)/$(BINARY_NAME) -file <video.mp4> [-loop] [-verbose]"

.PHONY: clean
clean:
	rm -f $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: help
help:
	@echo "streamtest - Video streaming test tool"
	@echo ""
	@echo "Targets:"
	@echo "  build  - Build the streamtest binary"
	@echo "  run    - Show usage information"
	@echo "  clean  - Remove binary"
	@echo "  help   - Show this help"
	@echo ""
	@echo "Usage:"
	@echo "  $(BINARY_DIR)/$(BINARY_NAME) -file video.mp4           # Play once"
	@echo "  $(BINARY_DIR)/$(BINARY_NAME) -file video.mp4 -loop     # Loop playback"
	@echo "  $(BINARY_DIR)/$(BINARY_NAME) -file video.mp4 -verbose  # Verbose output"
