#!/usr/bin/env bash
# Install script for asahi-map

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing asahi-map...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${YELLOW}Note: Installation requires root. Re-running with sudo...${NC}"
   exec sudo "$0" "$@"
fi

# Build
echo "Building asahi-map..."
go build -ldflags="-s -w" -o asahi-map ./cmd/asahi-map

# Install binary
echo "Installing binary to /usr/local/bin..."
install -m 755 asahi-map /usr/local/bin/asahi-map

# Create config directory
echo "Creating config directories..."
mkdir -p /etc/asahi-map/layouts

# Install config files
echo "Installing configuration files..."
cp configs/config.yaml /etc/asahi-map/
cp configs/layouts/*.yaml /etc/asahi-map/layouts/

# Create systemd user service
echo "Creating systemd user service..."
mkdir -p /usr/lib/systemd/user
cat > /usr/lib/systemd/user/asahi-map.service << 'EOF'
[Unit]
Description=Asahi-Map Option Key Mapper
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/asahi-map
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Setup udev rule for /dev/uinput access
echo "Setting up udev rules..."
cat > /etc/udev/rules.d/99-asahi-map.rules << 'EOF'
# Allow input group to access uinput
KERNEL=="uinput", GROUP="input", MODE="0660"
EOF

# Reload udev rules
udevadm control --reload-rules
udevadm trigger

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Add your user to the input group:"
echo "     sudo usermod -aG input \$USER"
echo ""
echo "  2. Log out and back in for group changes to take effect"
echo ""
echo "  3. Copy config to your home directory (optional):"
echo "     mkdir -p ~/.config/asahi-map/layouts"
echo "     cp /etc/asahi-map/config.yaml ~/.config/asahi-map/"
echo "     cp /etc/asahi-map/layouts/*.yaml ~/.config/asahi-map/layouts/"
echo ""
echo "  4. Start the service:"
echo "     systemctl --user enable --now asahi-map"
echo ""
echo "  Or run manually:"
echo "     asahi-map"
