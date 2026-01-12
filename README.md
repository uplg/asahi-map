# Asahi-Map

A lightweight Go application for handling macOS Option key shortcuts on Linux. Designed as a minimal alternative to Toshy, focusing solely on Option key special character mappings with minimal memory footprint.

## Features

- 🎹 **macOS-style Option key shortcuts** - Type special characters like on macOS
- 🌍 **Multiple keyboard layouts** - Support for AZERTY, QWERTY, and custom layouts
- 🖥️ **System tray integration** - Easy access via GTK system tray icon
- ⚡ **Lightweight** - Written in Go for minimal memory usage
- 🔧 **Configurable** - YAML-based configuration files
- 🔤 **Dead key support** - For accented characters (é, è, ê, etc.)

## Requirements

- Linux with evdev support
- GTK3 for system tray
- Root access or `input` group membership

## Installation

### From source

```bash
# Install dependencies (Fedora)
sudo dnf install gtk3-devel

# Build
go build -ldflags="-s -w" -o asahi-map ./cmd/asahi-map

# Install
sudo cp asahi-map /usr/local/bin/
sudo cp configs/*.yaml /etc/asahi-map/
```

### Setup permissions

```bash
# Add your user to the input group
sudo usermod -aG input $USER

# Logout and login again for changes to take effect
```

## Usage

```bash
# Run with default config
asahi-map

# Run with specific layout
asahi-map -layout azerty-mac

# Run with custom config directory
asahi-map -config /path/to/configs
```

## Configuration

Configuration files are in YAML format and located in `~/.config/asahi-map/` or `/etc/asahi-map/`.

### Main config (`config.yaml`)

```yaml
layout: azerty-mac
log_level: info
keyboard_device: auto  # or specific device path
```

### Layout files (`layouts/azerty-mac.yaml`)

```yaml
name: "AZERTY Mac"
description: "French AZERTY keyboard for Mac"

mappings:
  # Alt + key -> Unicode character
  alt:
    "a": "æ"
    "z": "Ω"
    # ...
  
  # Shift + Alt + key -> Unicode character  
  shift_alt:
    "a": "Æ"
    # ...
```

## Architecture

```
asahi-map/
├── cmd/asahi-map/      # Main entry point
├── internal/
│   ├── config/         # Configuration loading
│   ├── keyboard/       # evdev/uinput handling
│   ├── mappings/       # Key mapping logic
│   └── tray/           # System tray UI
└── configs/            # Default layout files
```

## Memory Usage

Typical memory usage: ~5-10 MB (compared to ~100+ MB for Toshy)

## License

MIT License

## Credits

Inspired by [Toshy](https://github.com/RedBearAK/toshy) but reimplemented in Go for efficiency.
