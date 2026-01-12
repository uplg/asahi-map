// Package keyboard handles evdev input and uinput output for key remapping.
package keyboard

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	evdev "github.com/holoplot/go-evdev"
)

type Device struct {
	path   string
	device *evdev.InputDevice
	name   string
}

// DeviceManager handles discovery and management of keyboard devices.
type DeviceManager struct {
	mu      sync.RWMutex
	devices map[string]*Device
	logger  *slog.Logger
}

func NewDeviceManager(logger *slog.Logger) *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]*Device),
		logger:  logger,
	}
}

// FindKeyboards discovers keyboard devices in /dev/input.
func (dm *DeviceManager) FindKeyboards() ([]*Device, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	pattern := "/dev/input/event*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("globbing input devices: %w", err)
	}

	var keyboards []*Device

	for _, path := range matches {
		dev, err := evdev.Open(path)
		if err != nil {
			dm.logger.Debug("cannot open device", "path", path, "error", err)
			continue
		}

		name, err := dev.Name()
		if err != nil {
			dev.Close()
			continue
		}

		// Check if device has key capabilities
		if !dm.isKeyboard(dev) {
			dev.Close()
			continue
		}

		device := &Device{
			path:   path,
			device: dev,
			name:   name,
		}

		// Skip virtual devices we might have created
		if strings.Contains(strings.ToLower(name), "asahi-map") {
			dev.Close()
			continue
		}

		dm.devices[path] = device
		keyboards = append(keyboards, device)

		dm.logger.Info("found keyboard", "name", name, "path", path)
	}

	return keyboards, nil
}

func (dm *DeviceManager) isKeyboard(dev *evdev.InputDevice) bool {
	// Check for EV_KEY capability
	capableTypes := dev.CapableTypes()
	for _, t := range capableTypes {
		if t == evdev.EV_KEY {
			// Check if it has typical keyboard keys
			keyCodes := dev.CapableEvents(evdev.EV_KEY)
			for _, code := range keyCodes {
				// Look for letter keys (KEY_A through KEY_Z)
				if code >= 30 && code <= 52 {
					return true
				}
			}
		}
	}
	return false
}

// GrabDevice takes exclusive control of a device.
func (dm *DeviceManager) GrabDevice(dev *Device) error {
	if err := dev.device.Grab(); err != nil {
		return fmt.Errorf("grabbing device %s: %w", dev.path, err)
	}
	dm.logger.Info("grabbed device", "name", dev.name)
	return nil
}

// ReleaseDevice releases exclusive control of a device.
func (dm *DeviceManager) ReleaseDevice(dev *Device) error {
	if err := dev.device.Ungrab(); err != nil {
		return fmt.Errorf("releasing device %s: %w", dev.path, err)
	}
	dm.logger.Info("released device", "name", dev.name)
	return nil
}

// Close closes all managed devices.
func (dm *DeviceManager) Close() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, dev := range dm.devices {
		dev.device.Close()
	}
	dm.devices = make(map[string]*Device)
}

// ReadEvents reads events from a device and sends them to a channel.
func ReadEvents(ctx context.Context, dev *Device, events chan<- *KeyEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ev, err := dev.device.ReadOne()
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("device disconnected: %s", dev.path)
				}
				return fmt.Errorf("reading event: %w", err)
			}

			// Only process key events
			if ev.Type == evdev.EV_KEY {
				keyEvent := &KeyEvent{
					Code:      uint16(ev.Code),
					Value:     ev.Value,
					Timestamp: ev.Time,
					Device:    dev,
				}
				events <- keyEvent
			}
		}
	}
}

func (d *Device) Path() string {
	return d.path
}

func (d *Device) Name() string {
	return d.name
}
