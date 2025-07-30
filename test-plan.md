# bs-image-stream-server Testing Plan

## Overview

This comprehensive testing plan ensures the bs-image-stream-server operates correctly across all target platforms and use cases. The plan includes automated testing tools, visual verification methods, and performance validation.

## Test Infrastructure

### Test Image Generation

**Purpose**: Create consistent, predictable test patterns for visual verification
**Location**: `test-images/` directory
**Generator**: `cmd/test-image-gen/main.go`

#### Test Image Series

1. **Vertical Bar Sweep Animation**
   - **Files**: `frame_001.jpg` to `frame_060.jpg` (60 frames)
   - **Resolution**: 1920x1080 (HD)
   - **Pattern**: White vertical bar (40px wide) sweeping left to right
   - **Movement**: 32 pixels per frame (completes sweep in 2 seconds at 30 FPS)
   - **Background**: Solid black (#000000)
   - **Bar Color**: Bright white (#FFFFFF)

2. **Static Reference Images**
   - `static_black.jpg` - Solid black reference
   - `static_white.jpg` - Solid white reference
   - `static_grid.jpg` - Grid pattern for alignment testing

3. **Color Gradient Series**
   - `gradient_001.jpg` to `gradient_030.jpg` (30 frames)
   - Horizontal color sweep from red to blue

4. **Frame Counter Series**
   - `counter_001.jpg` to `counter_120.jpg` (120 frames, 4 seconds)
   - Large frame number overlay for timing verification

### Test Image Generator

**File**: `cmd/test-image-gen/main.go`

```go
// Command-line interface:
// ./test-image-gen -pattern=sweep -frames=60 -width=1920 -height=1080
// ./test-image-gen -pattern=gradient -frames=30
// ./test-image-gen -pattern=counter -frames=120
```

**Features**:
- Multiple test pattern generation
- Configurable resolution and frame count
- Progress indicator during generation
- JPEG quality optimization for file size
- Automatic output directory creation

### Image Feeder Program

**File**: `cmd/test-feeder/main.go`

```go
// Command-line interface:
// ./test-feeder -source=test-images/frame_%03d.jpg -target=/tmp/output.jpg -fps=30
// ./test-feeder -source=test-images/counter_%03d.jpg -fps=60 -loop=false
```

**Features**:
- Precise 33ms timing (30 FPS) using `time.Ticker`
- Configurable frame rate for testing different speeds
- Loop control for continuous or single-pass operation
- Real-time statistics (actual FPS, timing accuracy)
- Graceful shutdown with Ctrl+C
- File copy error handling and recovery

## Test Categories

### 1. Unit Tests

**Location**: `internal/*/test.go` files
**Run with**: `make test`

#### File Monitor Tests
- `internal/monitor/file_monitor_test.go`
- Modification time detection accuracy
- Missing file handling
- File lock handling during copy operations
- Timer precision validation

#### Image Cache Tests
- `internal/cache/image_cache_test.go`
- Thread-safe access patterns
- Memory usage optimization
- ETag generation consistency
- Cache invalidation logic

#### HTTP Server Tests
- `internal/server/server_test.go`
- Route handling correctness
- HTTP header generation
- Error response handling
- Concurrent request processing

### 2. Integration Tests

**Location**: `test/integration/`
**Run with**: `make test-integration`

#### End-to-End Workflow
- Start image stream server
- Run test feeder program
- Verify HTTP responses contain expected images
- Validate ETag behavior
- Test graceful shutdown

#### Multi-Client Testing
- Simulate multiple browser connections
- Verify concurrent image delivery
- Test bandwidth utilization
- Validate cache effectiveness

### 3. Performance Tests

**Location**: `test/performance/`
**Run with**: `make test-performance`

#### Frame Rate Accuracy
- Measure actual vs. target 30 FPS delivery
- Test with various image sizes
- Monitor CPU and memory usage
- Validate under load conditions

#### Resource Usage Tests
- Memory consumption over time
- CPU utilization patterns
- Network bandwidth efficiency
- File I/O performance

#### Load Testing
- Multiple concurrent clients
- Sustained operation testing (24+ hours)
- Memory leak detection
- Performance degradation analysis

### 4. Visual Verification Tests

**Manual/Automated**: Browser-based validation
**Run with**: `make test-visual`

#### Animation Smoothness
- Visual inspection of bar sweep animation
- Frame dropping detection
- Timing consistency verification
- Browser compatibility testing

#### UI Responsiveness
- BrightSign branding display
- Responsive layout testing
- Error message display
- Loading state handling

### 5. Cross-Platform Tests

**Location**: `test/platforms/`
**Run with**: `make test-platforms`

#### Embedded Target Testing
- ARM (Raspberry Pi 3) validation
- ARM64 (Raspberry Pi 4) validation
- x86_64 Linux validation
- Performance comparison across platforms

#### Docker Environment Testing
- Container build verification
- Volume mounting functionality
- Network accessibility
- Resource constraints testing

## Test Execution Plan

### Phase 1: Development Testing

1. **Generate test images**:
   ```bash
   make init-project
   go run cmd/test-image-gen/main.go -pattern=sweep -frames=60
   go run cmd/test-image-gen/main.go -pattern=counter -frames=120
   ```

2. **Run unit tests**:
   ```bash
   make test
   make test-coverage
   ```

3. **Manual integration testing**:
   ```bash
   # Terminal 1: Start image stream server
   make run-debug
   
   # Terminal 2: Start test feeder
   go run cmd/test-feeder/main.go
   
   # Browser: Open http://localhost:8080
   ```

### Phase 2: Automated Testing

1. **Integration test suite**:
   ```bash
   make test-integration
   ```

2. **Performance validation**:
   ```bash
   make test-performance
   make load-test
   ```

3. **Cross-platform builds**:
   ```bash
   make build-embedded
   ```

### Phase 3: Production Readiness

1. **Extended load testing**:
   ```bash
   # 24-hour stability test
   make test-stability
   ```

2. **Embedded platform testing**:
   ```bash
   # Deploy to actual embedded devices
   make deploy-test-arm64
   ```

3. **Documentation validation**:
   ```bash
   # Verify all examples work
   make test-examples
   ```

## Test Automation

### Makefile Targets

```makefile
# Generate all test images
test-images:
	mkdir -p test-images
	go run cmd/test-image-gen/main.go -pattern=sweep -frames=60
	go run cmd/test-image-gen/main.go -pattern=gradient -frames=30
	go run cmd/test-image-gen/main.go -pattern=counter -frames=120

# Run complete test suite
test-full: test test-integration test-performance test-visual

# Visual verification test
test-visual: build test-images
	./bs-frame-monitor &
	sleep 2
	go run cmd/test-feeder/main.go -loop=false
	@echo "Open http://localhost:8080 to verify animation"
	pkill bs-frame-monitor

# Performance benchmarking
test-performance: build test-images
	./bs-frame-monitor &
	sleep 2
	go run test/performance/benchmark.go
	pkill bs-frame-monitor

# Cross-platform validation
test-platforms: build-embedded
	# Test each binary on appropriate platforms
	# (requires actual hardware or emulation)
```

### Continuous Integration

**GitHub Actions**: `.github/workflows/test.yml`
- Automated test execution on pull requests
- Cross-platform build verification
- Performance regression detection
- Test coverage reporting

## Success Criteria

### Functional Requirements
- [ ] 30 FPS image delivery verified with test animations
- [ ] Modification time detection working correctly
- [ ] Web interface displays properly with BrightSign branding
- [ ] ETag caching reduces bandwidth usage
- [ ] Graceful handling of missing/locked files

### Performance Requirements
- [ ] CPU usage < 10% on embedded ARM systems
- [ ] Memory usage < 50MB sustained operation
- [ ] Frame delivery timing within ±2ms of target
- [ ] Supports 10+ concurrent browser connections
- [ ] 24+ hour stability without memory leaks

### Platform Requirements
- [ ] Builds successfully for ARM, ARM64, x86_64
- [ ] Functions correctly on Raspberry Pi 3/4
- [ ] Docker development environment working
- [ ] All test tools functioning correctly

## Test Data Management

### Directory Structure
```
test-images/
├── sweep/
│   ├── frame_001.jpg
│   ├── frame_002.jpg
│   └── ... (60 files)
├── gradient/
│   ├── gradient_001.jpg
│   └── ... (30 files)
├── counter/
│   ├── counter_001.jpg
│   └── ... (120 files)
└── static/
    ├── black.jpg
    ├── white.jpg
    └── grid.jpg
```

### File Management
- Test images generated automatically
- Version control excludes generated files (`.gitignore`)
- Consistent naming convention
- Optimized JPEG quality for testing

This comprehensive testing plan ensures the bs-image-stream-server meets all functional and performance requirements while providing reliable tools for ongoing validation during development.