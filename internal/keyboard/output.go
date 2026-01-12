package keyboard

import (
	"fmt"
	"log/slog"

	"github.com/bendahl/uinput"
)

// VirtualKeyboard provides methods to inject key events and Unicode characters.
type VirtualKeyboard struct {
	keyboard uinput.Keyboard
	logger   *slog.Logger
}

// NewVirtualKeyboard creates a new virtual keyboard for output.
func NewVirtualKeyboard(logger *slog.Logger) (*VirtualKeyboard, error) {
	kb, err := uinput.CreateKeyboard("/dev/uinput", []byte("asahi-map-virtual"))
	if err != nil {
		return nil, fmt.Errorf("creating virtual keyboard: %w", err)
	}

	return &VirtualKeyboard{
		keyboard: kb,
		logger:   logger,
	}, nil
}

// Close releases the virtual keyboard.
func (vk *VirtualKeyboard) Close() error {
	return vk.keyboard.Close()
}

// PressKey simulates a key press.
func (vk *VirtualKeyboard) PressKey(code int) error {
	return vk.keyboard.KeyDown(code)
}

// ReleaseKey simulates a key release.
func (vk *VirtualKeyboard) ReleaseKey(code int) error {
	return vk.keyboard.KeyUp(code)
}

// TapKey simulates a key press and release.
func (vk *VirtualKeyboard) TapKey(code int) error {
	if err := vk.keyboard.KeyDown(code); err != nil {
		return err
	}
	return vk.keyboard.KeyUp(code)
}

// TypeUnicode types a Unicode character using the Ctrl+Shift+U method.
// This works in GTK/Qt applications that support Unicode input.
// On AZERTY keyboards, digits require Shift to be pressed.
func (vk *VirtualKeyboard) TypeUnicode(r rune) error {
	hex := fmt.Sprintf("%x", r) // lowercase hex

	vk.logger.Debug("typing unicode via ctrl+shift+u", "char", string(r), "hex", hex)

	// Press Ctrl+Shift+U
	if err := vk.keyboard.KeyDown(uinput.KeyLeftctrl); err != nil {
		return err
	}
	if err := vk.keyboard.KeyDown(uinput.KeyLeftshift); err != nil {
		vk.keyboard.KeyUp(uinput.KeyLeftctrl)
		return err
	}
	if err := vk.keyboard.KeyPress(uinput.KeyU); err != nil {
		vk.keyboard.KeyUp(uinput.KeyLeftshift)
		vk.keyboard.KeyUp(uinput.KeyLeftctrl)
		return err
	}
	if err := vk.keyboard.KeyUp(uinput.KeyLeftshift); err != nil {
		vk.keyboard.KeyUp(uinput.KeyLeftctrl)
		return err
	}
	if err := vk.keyboard.KeyUp(uinput.KeyLeftctrl); err != nil {
		return err
	}

	// Type hex digits - on AZERTY, digits need Shift
	for _, c := range hex {
		if err := vk.typeHexChar(c); err != nil {
			return err
		}
	}

	// Press Space to confirm
	if err := vk.keyboard.KeyPress(uinput.KeySpace); err != nil {
		return err
	}

	return nil
}

// typeHexChar types a single hex character (0-9, a-f).
// On AZERTY keyboards, digits require Shift to be pressed.
// Letters a-f are typed using their AZERTY physical positions.
func (vk *VirtualKeyboard) typeHexChar(c rune) error {
	switch c {
	// Digits 0-9: need Shift on AZERTY
	case '0':
		return vk.typeWithShift(uinput.Key0)
	case '1':
		return vk.typeWithShift(uinput.Key1)
	case '2':
		return vk.typeWithShift(uinput.Key2)
	case '3':
		return vk.typeWithShift(uinput.Key3)
	case '4':
		return vk.typeWithShift(uinput.Key4)
	case '5':
		return vk.typeWithShift(uinput.Key5)
	case '6':
		return vk.typeWithShift(uinput.Key6)
	case '7':
		return vk.typeWithShift(uinput.Key7)
	case '8':
		return vk.typeWithShift(uinput.Key8)
	case '9':
		return vk.typeWithShift(uinput.Key9)
	// Letters a-f: use AZERTY positions (KeyQ = 'a', KeyB = 'b', etc.)
	case 'a', 'A':
		return vk.keyboard.KeyPress(uinput.KeyQ) // 'a' is on Q key position on AZERTY
	case 'b', 'B':
		return vk.keyboard.KeyPress(uinput.KeyB)
	case 'c', 'C':
		return vk.keyboard.KeyPress(uinput.KeyC)
	case 'd', 'D':
		return vk.keyboard.KeyPress(uinput.KeyD)
	case 'e', 'E':
		return vk.keyboard.KeyPress(uinput.KeyE)
	case 'f', 'F':
		return vk.keyboard.KeyPress(uinput.KeyF)
	}
	return nil
}

// typeWithShift types a key with Shift held down.
func (vk *VirtualKeyboard) typeWithShift(keyCode int) error {
	if err := vk.keyboard.KeyDown(uinput.KeyLeftshift); err != nil {
		return err
	}
	if err := vk.keyboard.KeyPress(keyCode); err != nil {
		vk.keyboard.KeyUp(uinput.KeyLeftshift)
		return err
	}
	return vk.keyboard.KeyUp(uinput.KeyLeftshift)
}

// TypeString types a string character by character.
func (vk *VirtualKeyboard) TypeString(s string) error {
	for _, r := range s {
		if err := vk.TypeUnicode(r); err != nil {
			return err
		}
	}
	return nil
}

// PassthroughWithRAlt sends a key with Right Alt modifier.
func (vk *VirtualKeyboard) PassthroughWithRAlt(keyCode int) error {
	if err := vk.keyboard.KeyDown(uinput.KeyRightalt); err != nil {
		return err
	}
	if err := vk.keyboard.KeyPress(keyCode); err != nil {
		vk.keyboard.KeyUp(uinput.KeyRightalt)
		return err
	}
	return vk.keyboard.KeyUp(uinput.KeyRightalt)
}

// PassthroughWithShiftRAlt sends a key with Shift+Right Alt modifiers.
// shiftAlreadyDown indicates if Shift was already being held by the user.
// If true, we don't release Shift at the end to maintain the user's Shift state.
func (vk *VirtualKeyboard) PassthroughWithShiftRAlt(keyCode int, shiftAlreadyDown bool) error {
	// Only press Shift if it wasn't already down
	if !shiftAlreadyDown {
		if err := vk.keyboard.KeyDown(uinput.KeyLeftshift); err != nil {
			return err
		}
	}
	if err := vk.keyboard.KeyDown(uinput.KeyRightalt); err != nil {
		if !shiftAlreadyDown {
			vk.keyboard.KeyUp(uinput.KeyLeftshift)
		}
		return err
	}
	if err := vk.keyboard.KeyPress(keyCode); err != nil {
		vk.keyboard.KeyUp(uinput.KeyRightalt)
		if !shiftAlreadyDown {
			vk.keyboard.KeyUp(uinput.KeyLeftshift)
		}
		return err
	}
	if err := vk.keyboard.KeyUp(uinput.KeyRightalt); err != nil {
		if !shiftAlreadyDown {
			vk.keyboard.KeyUp(uinput.KeyLeftshift)
		}
		return err
	}
	// Only release Shift if we pressed it ourselves
	if !shiftAlreadyDown {
		return vk.keyboard.KeyUp(uinput.KeyLeftshift)
	}
	return nil
}

// ForwardEvent forwards an event unchanged.
func (vk *VirtualKeyboard) ForwardEvent(code uint16, value int32) error {
	switch value {
	case 0: // Release
		return vk.keyboard.KeyUp(int(code))
	case 1: // Press
		return vk.keyboard.KeyDown(int(code))
	case 2: // Repeat - send another key down (the kernel handles auto-repeat)
		// Note: We just send KeyDown again, not KeyPress (which would do Down+Up)
		// The key is already down, so another KeyDown triggers repeat in the kernel
		return vk.keyboard.KeyDown(int(code))
	}
	return nil
}
