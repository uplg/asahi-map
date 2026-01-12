package mappings

// KeyCode represents a Linux evdev key code.
type KeyCode uint16

// Common key codes from linux/input-event-codes.h
const (
	KEY_RESERVED   KeyCode = 0
	KEY_ESC        KeyCode = 1
	KEY_1          KeyCode = 2
	KEY_2          KeyCode = 3
	KEY_3          KeyCode = 4
	KEY_4          KeyCode = 5
	KEY_5          KeyCode = 6
	KEY_6          KeyCode = 7
	KEY_7          KeyCode = 8
	KEY_8          KeyCode = 9
	KEY_9          KeyCode = 10
	KEY_0          KeyCode = 11
	KEY_MINUS      KeyCode = 12
	KEY_EQUAL      KeyCode = 13
	KEY_BACKSPACE  KeyCode = 14
	KEY_TAB        KeyCode = 15
	KEY_Q          KeyCode = 16
	KEY_W          KeyCode = 17
	KEY_E          KeyCode = 18
	KEY_R          KeyCode = 19
	KEY_T          KeyCode = 20
	KEY_Y          KeyCode = 21
	KEY_U          KeyCode = 22
	KEY_I          KeyCode = 23
	KEY_O          KeyCode = 24
	KEY_P          KeyCode = 25
	KEY_LEFTBRACE  KeyCode = 26
	KEY_RIGHTBRACE KeyCode = 27
	KEY_ENTER      KeyCode = 28
	KEY_LEFTCTRL   KeyCode = 29
	KEY_A          KeyCode = 30
	KEY_S          KeyCode = 31
	KEY_D          KeyCode = 32
	KEY_F          KeyCode = 33
	KEY_G          KeyCode = 34
	KEY_H          KeyCode = 35
	KEY_J          KeyCode = 36
	KEY_K          KeyCode = 37
	KEY_L          KeyCode = 38
	KEY_SEMICOLON  KeyCode = 39
	KEY_APOSTROPHE KeyCode = 40
	KEY_GRAVE      KeyCode = 41
	KEY_LEFTSHIFT  KeyCode = 42
	KEY_BACKSLASH  KeyCode = 43
	KEY_Z          KeyCode = 44
	KEY_X          KeyCode = 45
	KEY_C          KeyCode = 46
	KEY_V          KeyCode = 47
	KEY_B          KeyCode = 48
	KEY_N          KeyCode = 49
	KEY_M          KeyCode = 50
	KEY_COMMA      KeyCode = 51
	KEY_DOT        KeyCode = 52
	KEY_SLASH      KeyCode = 53
	KEY_RIGHTSHIFT KeyCode = 54
	KEY_LEFTALT    KeyCode = 56
	KEY_SPACE      KeyCode = 57
	KEY_CAPSLOCK   KeyCode = 58
	KEY_102ND      KeyCode = 86
	KEY_RIGHTALT   KeyCode = 100
	KEY_LEFTMETA   KeyCode = 125
	KEY_RIGHTMETA  KeyCode = 126
)

// KeyCodeToName maps key codes to their string names (lowercase).
var KeyCodeToName = map[KeyCode]string{
	KEY_1:          "1",
	KEY_2:          "2",
	KEY_3:          "3",
	KEY_4:          "4",
	KEY_5:          "5",
	KEY_6:          "6",
	KEY_7:          "7",
	KEY_8:          "8",
	KEY_9:          "9",
	KEY_0:          "0",
	KEY_MINUS:      "minus",
	KEY_EQUAL:      "equal",
	KEY_Q:          "q",
	KEY_W:          "w",
	KEY_E:          "e",
	KEY_R:          "r",
	KEY_T:          "t",
	KEY_Y:          "y",
	KEY_U:          "u",
	KEY_I:          "i",
	KEY_O:          "o",
	KEY_P:          "p",
	KEY_LEFTBRACE:  "leftbrace",
	KEY_RIGHTBRACE: "rightbrace",
	KEY_A:          "a",
	KEY_S:          "s",
	KEY_D:          "d",
	KEY_F:          "f",
	KEY_G:          "g",
	KEY_H:          "h",
	KEY_J:          "j",
	KEY_K:          "k",
	KEY_L:          "l",
	KEY_SEMICOLON:  "semicolon",
	KEY_APOSTROPHE: "apostrophe",
	KEY_GRAVE:      "grave",
	KEY_BACKSLASH:  "backslash",
	KEY_Z:          "z",
	KEY_X:          "x",
	KEY_C:          "c",
	KEY_V:          "v",
	KEY_B:          "b",
	KEY_N:          "n",
	KEY_M:          "m",
	KEY_COMMA:      "comma",
	KEY_DOT:        "dot",
	KEY_SLASH:      "slash",
	KEY_SPACE:      "space",
	KEY_102ND:      "102nd",
}

// NameToKeyCode is the reverse mapping.
var NameToKeyCode map[string]KeyCode

func init() {
	NameToKeyCode = make(map[string]KeyCode, len(KeyCodeToName))
	for code, name := range KeyCodeToName {
		NameToKeyCode[name] = code
	}
}
