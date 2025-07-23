# BS Frame Monitor

A lightweight embedded Linux web server written in Go that provides real-time image streaming capabilities for digital signage monitoring applications. This project is part of the BrightSign Platform Modernization initiative.

## Overview

The BS Frame Monitor continuously monitors a local image file and serves it via HTTP at 30 FPS to web browsers, making it ideal for real-time visual monitoring of digital signage displays, camera feeds, or any application that generates periodic image updates.

### Key Features

- **High-Performance Streaming**: 30 FPS image updates with minimal CPU overhead
- **Embedded System Optimized**: Designed for resource-constrained Linux environments
- **Professional UI**: BrightSign-branded web interface with modern styling
- **Efficient Change Detection**: Uses file modification time instead of content hashing
- **Zero Dependencies**: Built with Go standard library only
- **Cross-Platform**: Supports ARM, ARM64, and x86_64 architectures

### Architecture

- **Web Server**: HTTP server serving static HTML and real-time image endpoints
- **File Monitor**: Efficient 30 FPS polling with modification time detection
- **Image Cache**: Memory-efficient caching with ETag support for bandwidth optimization
- **Professional Interface**: Responsive web UI with BrightSign branding

## Use Cases

- **Digital Signage Monitoring**: Real-time view of what's displayed on remote screens
- **Camera Feed Streaming**: Live monitoring of security or industrial cameras
- **Automated Testing**: Visual verification of display content in CI/CD pipelines
- **Remote Troubleshooting**: Quick visual assessment of remote display issues

## Development Environment

This project includes a secure Docker-based development environment that provides isolated execution of development tools while limiting access to only the current project directory.

### Prerequisites

- Docker installed and running
- Current user has Docker access permissions

### Building the Development Container

1. **Clone or navigate to the project directory**:
   ```bash
   cd bs-frame-monitor
   ```

2. **Build the development image**:
   ```bash
   docker build -t claude-dev-env .
   ```

3. **Verify the build**:
   ```bash
   docker run --rm claude-dev-env go version
   docker run --rm claude-dev-env node --version
   docker run --rm claude-dev-env npm --version
   ```

### Running the Development Environment

#### Interactive Development Session

Start an interactive shell with full development tools:

```bash
docker run -it --rm -v $(pwd):/workspace claude-dev-env
```

This provides access to:
- Go development environment
- Node.js 20.x with npm and pnpm
- TypeScript compiler
- Essential tools: git, curl, wget, vim, nano
- Claude CLI with convenient `clauded` alias

#### Using Claude Code Safely

The development environment includes a pre-configured alias for safe Claude Code execution:

```bash
# Inside the container
clauded  # Equivalent to: claude --dangerously-skip-permissions
```

#### Direct Claude Execution

Run Claude Code directly without interactive session:

```bash
docker run -it --rm -v $(pwd):/workspace claude-dev-env bash -c "cd /workspace && claude --dangerously-skip-permissions"
```

#### Convenience Script

Create a wrapper script for easier access:

```bash
#!/bin/bash
# claude-safe.sh
docker run -it --rm -v $(pwd):/workspace claude-dev-env bash -c "cd /workspace && claude --dangerously-skip-permissions $*"
```

Make it executable:
```bash
chmod +x claude-safe.sh
./claude-safe.sh
```

### Security Benefits

The Docker environment provides:

- **Filesystem Isolation**: Access limited to current project directory only
- **Process Isolation**: Container-level separation from host system
- **Easy Cleanup**: Automatic container removal with `--rm` flag
- **Permission Boundaries**: Even with `--dangerously-skip-permissions`, damage is contained
- **No System Access**: Cannot modify host system files or other projects

### Development Workflow

1. Navigate to your project directory
2. Start the development environment with Docker
3. Use the `clauded` alias for convenient Claude Code execution
4. Develop normally with all tools available
5. Exit safely - container auto-removes

### Container Management

```bash
# List running containers
docker ps

# Stop all containers
docker stop $(docker ps -q)

# Remove unused images
docker image prune

# Complete cleanup
docker system prune -a
```

## Project Structure

```
bs-frame-monitor/
├── CLAUDE.md              # Claude Code guidance and implementation plan
├── docker-plan.md         # Detailed Docker development environment plan
├── Dockerfile             # Development environment container
├── README.md              # This file
└── (implementation files to be created)
    ├── main.go            # Application entry point
    ├── internal/
    │   ├── server/        # HTTP server implementation
    │   ├── monitor/       # File monitoring logic
    │   └── cache/         # Image caching system
    └── web/
        └── index.html     # BrightSign-branded web interface
```

## Next Steps

1. **Set up development environment** using the Docker instructions above
2. **Review implementation plan** in `CLAUDE.md`
3. **Start development** using the `clauded` command in the secure container
4. **Build the Go application** following the architecture outlined in the plan

For detailed implementation guidance, see `CLAUDE.md`. For Docker environment details, see `docker-plan.md`.