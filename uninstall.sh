#!/usr/bin/env bash
# Uninstall script for asahi-map

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}Uninstalling asahi-map...${NC}"

# Get the actual user (even when running with sudo)
if [[ -n "$SUDO_USER" ]]; then
    REAL_USER="$SUDO_USER"
    REAL_HOME=$(getent passwd "$SUDO_USER" | cut -d: -f6)
else
    REAL_USER="$USER"
    REAL_HOME="$HOME"
fi

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${YELLOW}Note: Uninstallation requires root. Re-running with sudo...${NC}"
   exec sudo "$0" "$@"
fi

# Stop running instances
echo "Stopping asahi-map if running..."
pkill -f asahi-map 2>/dev/null || true

# Disable and stop systemd user service (as the real user)
echo "Disabling systemd user service..."
if [[ -f /usr/lib/systemd/user/asahi-map.service ]]; then
    sudo -u "$REAL_USER" systemctl --user stop asahi-map.service 2>/dev/null || true
    sudo -u "$REAL_USER" systemctl --user disable asahi-map.service 2>/dev/null || true
fi

# Remove binary
echo "Removing binary..."
rm -f /usr/local/bin/asahi-map

# Remove systemd user service
echo "Removing systemd user service..."
rm -f /usr/lib/systemd/user/asahi-map.service

# Remove autostart entry
echo "Removing autostart entry..."
rm -f /etc/xdg/autostart/asahi-map.desktop

# Remove udev rules
echo "Removing udev rules..."
rm -f /etc/udev/rules.d/99-asahi-map.rules

# Reload udev rules
udevadm control --reload-rules 2>/dev/null || true
udevadm trigger 2>/dev/null || true

# Ask about config files
echo ""
echo -e "${YELLOW}Configuration files found:${NC}"

CONFIG_EXISTS=false
if [[ -d /etc/asahi-map ]]; then
    echo "  - /etc/asahi-map/"
    CONFIG_EXISTS=true
fi

USER_CONFIG_DIR="$REAL_HOME/.config/asahi-map"
if [[ -d "$USER_CONFIG_DIR" ]]; then
    echo "  - $USER_CONFIG_DIR/"
    CONFIG_EXISTS=true
fi

if [[ "$CONFIG_EXISTS" == "true" ]]; then
    echo ""
    read -p "Do you want to remove configuration files? [y/N]: " remove_config
    case "$remove_config" in
        y|Y|yes|YES)
            echo "Removing configuration files..."
            rm -rf /etc/asahi-map
            rm -rf "$USER_CONFIG_DIR"
            echo -e "  ${GREEN}[REMOVED]${NC} Configuration files deleted"
            ;;
        *)
            echo -e "  ${YELLOW}[KEPT]${NC} Configuration files preserved"
            ;;
    esac
else
    echo "  (none found)"
fi

echo ""
echo -e "${GREEN}Uninstallation complete!${NC}"
echo ""
echo "Note: If you added your user to the 'input' group for asahi-map,"
echo "you may remove them with: sudo gpasswd -d \$USER input"
