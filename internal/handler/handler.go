// Package handler coordinates keyboard input processing and key mapping.
package handler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/leonard/asahi-map/internal/keyboard"
	"github.com/leonard/asahi-map/internal/mappings"
)

// Handler processes keyboard events and applies mappings.
type Handler struct {
	mu            sync.RWMutex
	lookup        *mappings.KeyLookup
	vkb           *keyboard.VirtualKeyboard
	keyState      *keyboard.KeyState
	enabled       bool
	logger        *slog.Logger

	// Track keys we've intercepted to properly handle release
	interceptedKeys map[uint16]bool
}

// New creates a new keyboard event handler.
func New(lookup *mappings.KeyLookup, vkb *keyboard.VirtualKeyboard, logger *slog.Logger) *Handler {
	return &Handler{
		lookup:          lookup,
		vkb:             vkb,
		keyState:        &keyboard.KeyState{},
		enabled:         true,
		logger:          logger,
		interceptedKeys: make(map[uint16]bool),
	}
}

// SetEnabled enables or disables key mapping.
func (h *Handler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = enabled
	h.logger.Info("handler state changed", "enabled", enabled)
}

// SetLayout changes the active key mapping layout.
func (h *Handler) SetLayout(lookup *mappings.KeyLookup) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lookup = lookup
	h.logger.Info("layout changed")
}

// ProcessEvents reads events from the channel and processes them.
func (h *Handler) ProcessEvents(ctx context.Context, events <-chan *keyboard.KeyEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev := <-events:
			if err := h.handleEvent(ev); err != nil {
				h.logger.Error("error handling event", "error", err)
			}
		}
	}
}

// handleEvent processes a single key event.
func (h *Handler) handleEvent(ev *keyboard.KeyEvent) error {
	// Always update modifier state
	h.keyState.UpdateFromEvent(ev)

	// Debug logging for all key events
	keyName, hasName := mappings.KeyCodeToName[mappings.KeyCode(ev.Code)]
	if !hasName {
		keyName = "unknown"
	}
	h.logger.Debug("key event",
		"code", ev.Code,
		"key", keyName,
		"value", ev.Value,
		"leftAlt", h.keyState.LeftAltPressed(),
		"shift", h.keyState.ShiftPressed(),
	)

	// IMPORTANT: Don't forward Left Alt at all - we consume it entirely
	// This prevents KDE/GTK/Qt from showing menus when Alt is pressed
	// Users can still use Right Alt for system shortcuts
	if ev.Code == keyboard.KEY_LEFTALT {
		h.logger.Debug("consuming left alt (not forwarding)")
		return nil
	}

	// Forward other modifier keys normally
	if keyboard.IsModifier(ev.Code) {
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	h.mu.RLock()
	enabled := h.enabled
	lookup := h.lookup
	h.mu.RUnlock()

	// If disabled, just forward the event
	if !enabled {
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Handle key release for previously intercepted keys
	if ev.IsRelease() {
		h.mu.Lock()
		wasIntercepted := h.interceptedKeys[ev.Code]
		delete(h.interceptedKeys, ev.Code)
		h.mu.Unlock()

		if wasIntercepted {
			// We already handled this key, don't forward the release
			return nil
		}
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Only process on key press (not repeat for now)
	if !ev.IsPress() {
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Check if Left Alt (Option) is pressed
	if !h.keyState.LeftAltPressed() {
		// Check for dead key processing
		if h.lookup.HasActiveDeadKey() {
			return h.handleDeadKeyCombo(ev)
		}
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Get the key name
	keyName, ok := mappings.KeyCodeToName[mappings.KeyCode(ev.Code)]
	if !ok {
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Look up the mapping
	var mapping *mappings.Mapping
	if h.keyState.ShiftPressed() {
		mapping = lookup.LookupShiftAlt(keyName)
	} else {
		mapping = lookup.LookupAlt(keyName)
	}

	if mapping == nil {
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	// Mark this key as intercepted
	h.mu.Lock()
	h.interceptedKeys[ev.Code] = true
	h.mu.Unlock()

	// Handle the mapping
	return h.executeMapping(mapping, ev.Code)
}

// executeMapping executes a key mapping.
func (h *Handler) executeMapping(m *mappings.Mapping, keyCode uint16) error {
	// Handle passthrough (e.g., Alt-5 -> RAlt-5 for {)
	if m.Passthrough != "" {
		passthroughCode, ok := mappings.NameToKeyCode[m.Passthrough]
		if !ok {
			h.logger.Warn("unknown passthrough key", "key", m.Passthrough)
			return nil
		}
		h.logger.Debug("passthrough", "from", keyCode, "to", m.Passthrough, "toCode", passthroughCode, "shift", h.keyState.ShiftPressed())
		if h.keyState.ShiftPressed() {
			return h.vkb.PassthroughWithShiftRAlt(int(passthroughCode))
		}
		return h.vkb.PassthroughWithRAlt(int(passthroughCode))
	}

	// Handle dead key
	if m.IsDeadKey {
		h.lookup.SetDeadKey(m.DeadKeyID)
		// Also output the base accent character
		if r, ok := m.GetOutput(); ok {
			return h.vkb.TypeUnicode(r)
		}
		return nil
	}

	// Handle Unicode character
	if r, ok := m.GetOutput(); ok {
		h.logger.Debug("typing unicode", "char", string(r), "codepoint", r)
		return h.vkb.TypeUnicode(r)
	}

	return nil
}

// handleDeadKeyCombo processes a key after a dead key.
func (h *Handler) handleDeadKeyCombo(ev *keyboard.KeyEvent) error {
	keyName, ok := mappings.KeyCodeToName[mappings.KeyCode(ev.Code)]
	if !ok {
		h.lookup.ClearDeadKey()
		return h.vkb.ForwardEvent(ev.Code, ev.Value)
	}

	result, applied := h.lookup.ApplyDeadKey(keyName)
	if applied {
		h.mu.Lock()
		h.interceptedKeys[ev.Code] = true
		h.mu.Unlock()
		return h.vkb.TypeString(result)
	}

	return h.vkb.ForwardEvent(ev.Code, ev.Value)
}
