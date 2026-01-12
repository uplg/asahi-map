// Package mappings defines key mapping structures and Unicode output handling.
package mappings

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Layout represents a keyboard layout with Option key mappings.
type Layout struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Alt key mappings: key -> unicode codepoint or string
	Alt map[string]Mapping `yaml:"alt"`

	// Shift+Alt key mappings
	ShiftAlt map[string]Mapping `yaml:"shift_alt"`

	// Dead keys for accented characters
	DeadKeys map[string]DeadKey `yaml:"dead_keys"`
}

// Mapping represents a single key mapping.
type Mapping struct {
	// Output can be a single Unicode character or codepoint
	Char      string `yaml:"char,omitempty"`
	Codepoint uint32 `yaml:"codepoint,omitempty"`

	// For dead keys
	IsDeadKey bool   `yaml:"dead_key,omitempty"`
	DeadKeyID string `yaml:"dead_key_id,omitempty"`

	// For key pass-through (e.g., Alt-5 -> RAlt-5 for {)
	Passthrough string `yaml:"passthrough,omitempty"`
}

// DeadKey represents a dead key accent that combines with the next character.
type DeadKey struct {
	// Base accent character (shown when followed by space)
	Base string `yaml:"base"`

	// Combinations: base letter -> accented letter
	Combinations map[string]string `yaml:"combinations"`
}

// GetOutput returns the Unicode character or codepoint for this mapping.
func (m *Mapping) GetOutput() (rune, bool) {
	if m.Codepoint != 0 {
		return rune(m.Codepoint), true
	}
	if m.Char != "" {
		runes := []rune(m.Char)
		if len(runes) > 0 {
			return runes[0], true
		}
	}
	return 0, false
}

// LoadLayout reads a layout file from disk.
func LoadLayout(path string) (*Layout, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading layout file: %w", err)
	}

	var layout Layout
	if err := yaml.Unmarshal(data, &layout); err != nil {
		return nil, fmt.Errorf("parsing layout file: %w", err)
	}

	return &layout, nil
}

// KeyLookup provides efficient key mapping lookups.
type KeyLookup struct {
	layout        *Layout
	altMap        map[string]*Mapping
	shiftAltMap   map[string]*Mapping
	activeDeadKey *DeadKey
}

// NewKeyLookup creates a new key lookup from a layout.
func NewKeyLookup(layout *Layout) *KeyLookup {
	kl := &KeyLookup{
		layout:      layout,
		altMap:      make(map[string]*Mapping),
		shiftAltMap: make(map[string]*Mapping),
	}

	// Build lookup maps for O(1) access
	for k, v := range layout.Alt {
		mapping := v // Create copy to avoid pointer issues
		kl.altMap[k] = &mapping
	}
	for k, v := range layout.ShiftAlt {
		mapping := v
		kl.shiftAltMap[k] = &mapping
	}

	return kl
}

// LookupAlt returns the mapping for Alt+key.
func (kl *KeyLookup) LookupAlt(key string) *Mapping {
	return kl.altMap[key]
}

// LookupShiftAlt returns the mapping for Shift+Alt+key.
func (kl *KeyLookup) LookupShiftAlt(key string) *Mapping {
	return kl.shiftAltMap[key]
}

// SetDeadKey activates a dead key for the next character.
func (kl *KeyLookup) SetDeadKey(id string) {
	if dk, ok := kl.layout.DeadKeys[id]; ok {
		kl.activeDeadKey = &dk
	}
}

// ClearDeadKey clears the active dead key.
func (kl *KeyLookup) ClearDeadKey() {
	kl.activeDeadKey = nil
}

// HasActiveDeadKey returns true if a dead key is active.
func (kl *KeyLookup) HasActiveDeadKey() bool {
	return kl.activeDeadKey != nil
}

// ApplyDeadKey attempts to combine the active dead key with a character.
// Returns the combined character, or the base accent if no combination exists.
func (kl *KeyLookup) ApplyDeadKey(char string) (string, bool) {
	if kl.activeDeadKey == nil {
		return char, false
	}

	dk := kl.activeDeadKey
	kl.activeDeadKey = nil

	if combined, ok := dk.Combinations[char]; ok {
		return combined, true
	}

	// No combination found, return accent + original char
	return dk.Base + char, true
}
