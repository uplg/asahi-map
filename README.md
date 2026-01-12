# Asahi-Map

A lightweight Go application for handling macOS Option key shortcuts on Linux. Designed as a minimal alternative to Toshy, focusing solely on Option key special character mappings with minimal memory footprint.

## Features

- 🎹 **macOS-style Option key shortcuts** - Type special characters like on macOS
- 🇫🇷 **AZERTY Mac layout** - Full French keyboard support with Option and Option+Shift mappings
- 🖥️ **System tray integration** - Status icon and layout switching via fyne.io/systray
- ⚡ **Lightweight** - Written in Go for minimal memory usage (~5 MB)
- 🔧 **Configurable** - YAML-based configuration files
- 🔄 **AltGr passthrough** - Works on Wayland, X11, and all applications

## Requirements

- Linux with evdev support (Asahi Linux / Fedora)
- Root access or `input` group membership

## Installation

### From source

```bash
# Build
go build -ldflags="-s -w" -o asahi-map ./cmd/asahi-map

# Install
sudo ./install.sh
```

### Setup permissions

```bash
# Add your user to the input group
sudo usermod -aG input $USER

# Logout and login again for changes to take effect
```

## Usage

```bash
# Run with system tray (default)
asahi-map

# Run without system tray
asahi-map -no-tray

# Run with specific layout
asahi-map -layout azerty-mac

# Show version
asahi-map -version
```

## Configuration

Configuration files are in YAML format and located in `/etc/asahi-map/`.

### Layout file (`layouts/azerty-mac.yaml`)

Mappings use AltGr passthrough for maximum compatibility:

```yaml
name: "AZERTY Mac"
description: "French AZERTY keyboard for Mac"

alt:
  "q":  # Option+A (AZERTY)
    passthrough: "q"  # sends AltGr+Q → æ
  
shift_alt:
  "q":  # Option+Shift+A
    passthrough: "q"  # sends AltGr+Shift+Q → Æ
```

## License

MIT License

## Credits

Inspired by [Toshy](https://github.com/RedBearAK/toshy) but reimplemented in Go for efficiency.
