# Asahi-Map Project Instructions

## Project Overview
Asahi-Map is a lightweight Go application for handling macOS Option key shortcuts on Linux (specifically for Asahi Linux on Apple Silicon Macs).

## Architecture
- **cmd/asahi-map/** - Main application entry point
- **internal/config/** - Configuration management (YAML)
- **internal/keyboard/** - evdev/uinput keyboard handling
- **internal/mappings/** - Key mapping definitions and Unicode output
- **internal/tray/** - System tray icon and menu (GTK)
- **configs/** - Layout configuration files (AZERTY, QWERTY, etc.)

## Development Guidelines
- Use Go modules for dependency management
- Follow DRY principles - centralize common functionality
- Keep memory footprint minimal - avoid unnecessary allocations
- Use structured logging with slog
- Handle errors explicitly, no panic in library code
- Use interfaces for testability

## Key Technologies
- **evdev** - Reading keyboard input events
- **uinput** - Injecting synthetic key events
- **GTK3** - System tray icon via gotk3
- **YAML** - Configuration files

## Building
```bash
go build -ldflags="-s -w" -o asahi-map ./cmd/asahi-map
```

## Running
Requires root or input group membership for /dev/input access.
