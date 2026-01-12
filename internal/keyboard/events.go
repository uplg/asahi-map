package keyboard

import (
	"syscall"
)

type KeyEvent struct {
	Code      uint16
	Value     int32 // 0=release, 1=press, 2=repeat
	Timestamp syscall.Timeval
	Device    *Device
}

func (e *KeyEvent) IsPress() bool {
	return e.Value == 1
}

func (e *KeyEvent) IsRelease() bool {
	return e.Value == 0
}

func (e *KeyEvent) IsRepeat() bool {
	return e.Value == 2
}

type KeyState struct {
	LeftAlt    bool
	RightAlt   bool
	LeftShift  bool
	RightShift bool
	LeftCtrl   bool
	RightCtrl  bool
	LeftMeta   bool
	RightMeta  bool
}

const (
	KEY_LEFTSHIFT  uint16 = 42
	KEY_RIGHTSHIFT uint16 = 54
	KEY_LEFTCTRL   uint16 = 29
	KEY_RIGHTCTRL  uint16 = 97
	KEY_LEFTALT    uint16 = 56
	KEY_RIGHTALT   uint16 = 100
	KEY_LEFTMETA   uint16 = 125
	KEY_RIGHTMETA  uint16 = 126
)

func (ks *KeyState) UpdateFromEvent(ev *KeyEvent) {
	pressed := ev.IsPress()
	released := ev.IsRelease()

	switch ev.Code {
	case KEY_LEFTALT:
		if pressed {
			ks.LeftAlt = true
		} else if released {
			ks.LeftAlt = false
		}
	case KEY_RIGHTALT:
		if pressed {
			ks.RightAlt = true
		} else if released {
			ks.RightAlt = false
		}
	case KEY_LEFTSHIFT:
		if pressed {
			ks.LeftShift = true
		} else if released {
			ks.LeftShift = false
		}
	case KEY_RIGHTSHIFT:
		if pressed {
			ks.RightShift = true
		} else if released {
			ks.RightShift = false
		}
	case KEY_LEFTCTRL:
		if pressed {
			ks.LeftCtrl = true
		} else if released {
			ks.LeftCtrl = false
		}
	case KEY_RIGHTCTRL:
		if pressed {
			ks.RightCtrl = true
		} else if released {
			ks.RightCtrl = false
		}
	case KEY_LEFTMETA:
		if pressed {
			ks.LeftMeta = true
		} else if released {
			ks.LeftMeta = false
		}
	case KEY_RIGHTMETA:
		if pressed {
			ks.RightMeta = true
		} else if released {
			ks.RightMeta = false
		}
	}
}

func (ks *KeyState) AltPressed() bool {
	return ks.LeftAlt || ks.RightAlt
}

func (ks *KeyState) LeftAltPressed() bool {
	return ks.LeftAlt
}

func (ks *KeyState) ShiftPressed() bool {
	return ks.LeftShift || ks.RightShift
}

func (ks *KeyState) CtrlPressed() bool {
	return ks.LeftCtrl || ks.RightCtrl
}

func (ks *KeyState) MetaPressed() bool {
	return ks.LeftMeta || ks.RightMeta
}

func IsModifier(code uint16) bool {
	switch code {
	case KEY_LEFTALT, KEY_RIGHTALT,
		KEY_LEFTSHIFT, KEY_RIGHTSHIFT,
		KEY_LEFTCTRL, KEY_RIGHTCTRL,
		KEY_LEFTMETA, KEY_RIGHTMETA:
		return true
	}
	return false
}
