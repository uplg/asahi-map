#!/usr/bin/env bash
# Install script for asahi-map

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing asahi-map...${NC}"

# Get file hash (sha256, first 16 chars for readability)
file_hash() {
    if [[ -f "$1" ]]; then
        sha256sum "$1" 2>/dev/null | cut -c1-16
    else
        echo ""
    fi
}

# Compare source config with all installed locations
# Returns: "identical" if all match, "different" if any differ, "new" if no installed config exists
compare_configs() {
    local src_file="$1"
    local system_file="$2"
    local user_file="$3"
    
    local src_hash=$(file_hash "$src_file")
    local system_hash=$(file_hash "$system_file")
    local user_hash=$(file_hash "$user_file")
    
    # If no installed config exists anywhere
    if [[ -z "$system_hash" && -z "$user_hash" ]]; then
        echo "new"
        return
    fi
    
    # Check if any installed config differs from source
    if [[ -n "$system_hash" && "$system_hash" != "$src_hash" ]]; then
        echo "different"
        return
    fi
    
    if [[ -n "$user_hash" && "$user_hash" != "$src_hash" ]]; then
        echo "different"
        return
    fi
    
    echo "identical"
}

# Prompt user for config update choice
# Returns: "replace", "skip", or "manual"
prompt_config_action() {
    local config_name="$1"
    local locations="$2"
    
    # All display messages go to stderr, only return value goes to stdout
    echo "" >&2
    echo -e "${YELLOW}Configuration change detected for: ${config_name}${NC}" >&2
    echo "Installed locations with different content:" >&2
    echo "$locations" >&2
    echo "" >&2
    echo "What would you like to do?" >&2
    echo "  [r] Replace with new configuration (your changes will be lost)" >&2
    echo "  [s] Skip - keep your current configuration" >&2
    echo "  [m] Manual - I'll update it myself later" >&2
    echo "" >&2
    
    while true; do
        read -p "Your choice [r/s/m]: " choice
        case "$choice" in
            r|R) echo "replace"; return ;;
            s|S) echo "skip"; return ;;
            m|M) echo "manual"; return ;;
            *) echo "Invalid choice. Please enter r, s, or m." >&2 ;;
        esac
    done
}

# Install a config file with comparison check
# Args: source_file, system_dest, user_dest, config_display_name
install_config_with_check() {
    local src="$1"
    local system_dest="$2"
    local user_dest="$3"
    local display_name="$4"
    
    local status=$(compare_configs "$src" "$system_dest" "$user_dest")
    
    case "$status" in
        "new")
            # No existing config, just install
            mkdir -p "$(dirname "$system_dest")"
            cp "$src" "$system_dest"
            echo -e "  ${GREEN}[NEW]${NC} $display_name -> $system_dest"
            ;;
        "identical")
            # Same content, no action needed
            echo -e "  ${GREEN}[OK]${NC} $display_name (unchanged)"
            ;;
        "different")
            # Build list of differing locations
            local diff_locations=""
            local src_hash=$(file_hash "$src")
            
            if [[ -f "$system_dest" ]]; then
                local sys_hash=$(file_hash "$system_dest")
                if [[ "$sys_hash" != "$src_hash" ]]; then
                    diff_locations+="    - $system_dest\n"
                fi
            fi
            
            if [[ -f "$user_dest" ]]; then
                local usr_hash=$(file_hash "$user_dest")
                if [[ "$usr_hash" != "$src_hash" ]]; then
                    diff_locations+="    - $user_dest\n"
                fi
            fi
            
            local action=$(prompt_config_action "$display_name" "$(echo -e "$diff_locations")")
            
            case "$action" in
                "replace")
                    mkdir -p "$(dirname "$system_dest")"
                    cp "$src" "$system_dest"
                    echo -e "  ${GREEN}[REPLACED]${NC} $display_name -> $system_dest"
                    
                    # Also update user config if it exists and differs
                    if [[ -f "$user_dest" ]]; then
                        local usr_hash=$(file_hash "$user_dest")
                        if [[ "$usr_hash" != "$src_hash" ]]; then
                            cp "$src" "$user_dest"
                            echo -e "  ${GREEN}[REPLACED]${NC} $display_name -> $user_dest"
                        fi
                    fi
                    ;;
                "skip")
                    echo -e "  ${YELLOW}[SKIPPED]${NC} $display_name (keeping current config)"
                    ;;
                "manual")
                    echo -e "  ${YELLOW}[MANUAL]${NC} $display_name (update it yourself later)"
                    echo "    New config available at: $src"
                    ;;
            esac
            ;;
    esac
}

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
   echo -e "${YELLOW}Note: Installation requires root. Re-running with sudo...${NC}"
   exec sudo "$0" "$@"
fi

# Build
echo "Building asahi-map..."
go build -ldflags="-s -w" -o asahi-map ./cmd/asahi-map

# Install binary
echo "Installing binary to /usr/local/bin..."
install -m 755 asahi-map /usr/local/bin/asahi-map

# Create config directories
echo "Creating config directories..."
mkdir -p /etc/asahi-map/layouts

# User config directory
USER_CONFIG_DIR="$REAL_HOME/.config/asahi-map"

# Install config files with change detection
echo "Checking configuration files..."

# Main config file
install_config_with_check \
    "configs/config.yaml" \
    "/etc/asahi-map/config.yaml" \
    "$USER_CONFIG_DIR/config.yaml" \
    "config.yaml"

# Layout files
for layout in configs/layouts/*.yaml; do
    layout_name=$(basename "$layout")
    install_config_with_check \
        "$layout" \
        "/etc/asahi-map/layouts/$layout_name" \
        "$USER_CONFIG_DIR/layouts/$layout_name" \
        "layouts/$layout_name"
done

# Create systemd user service (without tray, for headless/service mode)
echo "Creating systemd user service..."
mkdir -p /usr/lib/systemd/user
cat > /usr/lib/systemd/user/asahi-map.service << 'EOF'
[Unit]
Description=Asahi-Map Option Key Mapper
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/asahi-map --no-tray
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Create XDG autostart entry (with tray, for desktop use)
echo "Creating autostart entry..."
mkdir -p /etc/xdg/autostart
cat > /etc/xdg/autostart/asahi-map.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=Asahi-Map
Comment=macOS Option key shortcuts for Linux
Exec=/usr/local/bin/asahi-map
Icon=input-keyboard
Terminal=false
Categories=Utility;
X-GNOME-Autostart-enabled=true
X-KDE-autostart-after=panel
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
echo "  1. Add your user to the input group (if not already done):"
echo "     sudo usermod -aG input \$USER"
echo ""
echo "  2. Log out and back in for group membership to take effect"
echo ""
echo "  3. Asahi-map will start automatically on next login (with systray)"
echo ""
echo "  Or run manually now (after logout/login):"
echo "     asahi-map"