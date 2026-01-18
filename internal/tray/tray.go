// Package tray provides system tray integration using fyne.io/systray.
package tray

import (
	"log/slog"

	"fyne.io/systray"
)

// Tray represents the system tray icon and menu.
type Tray struct {
	logger *slog.Logger

	// Callbacks
	onLayoutChange func(layout string)
	onToggle       func(enabled bool)
	onQuit         func()

	// State
	enabled          bool
	currentLayout    string
	availableLayouts []string

	// Menu items for updates
	statusItem  *systray.MenuItem
	layoutMenu  *systray.MenuItem
	layoutItems []*systray.MenuItem
}

// Config holds tray configuration.
type Config struct {
	CurrentLayout    string
	AvailableLayouts []string
	Enabled          bool
	OnLayoutChange   func(layout string)
	OnToggle         func(enabled bool)
	OnQuit           func()
	Logger           *slog.Logger
}

func New(cfg Config) *Tray {
	return &Tray{
		enabled:          cfg.Enabled,
		currentLayout:    cfg.CurrentLayout,
		availableLayouts: cfg.AvailableLayouts,
		onLayoutChange:   cfg.OnLayoutChange,
		onToggle:         cfg.OnToggle,
		onQuit:           cfg.OnQuit,
		logger:           cfg.Logger,
	}
}

// Run starts the system tray. This blocks until Quit is called.
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

// onReady is called when systray is ready.
func (t *Tray) onReady() {
	systray.SetIcon(keyboardIcon)
	systray.SetTitle("Asahi-Map")
	t.updateTooltip()

	// Status toggle
	t.statusItem = systray.AddMenuItem("✓ Enabled", "Toggle Option key mapping")

	systray.AddSeparator()

	// Layout submenu
	t.layoutMenu = systray.AddMenuItem(t.currentLayout+"    ", "Select keyboard layout")
	t.layoutItems = make([]*systray.MenuItem, len(t.availableLayouts))

	for i, layout := range t.availableLayouts {
		t.layoutItems[i] = t.layoutMenu.AddSubMenuItem(layout, "Switch to "+layout)
		if layout == t.currentLayout {
			t.layoutItems[i].Check()
		}
	}

	systray.AddSeparator()

	// Quit
	quitItem := systray.AddMenuItem("Quit", "Exit Asahi-Map")

	// Handle menu clicks
	go t.handleClicks(quitItem)
}

// handleClicks processes menu item clicks.
func (t *Tray) handleClicks(quitItem *systray.MenuItem) {
	// Handle status toggle
	go func() {
		for range t.statusItem.ClickedCh {
			t.toggleEnabled()
		}
	}()

	// Handle layout items
	for i, item := range t.layoutItems {
		go func(idx int, menuItem *systray.MenuItem) {
			for range menuItem.ClickedCh {
				t.selectLayout(t.availableLayouts[idx])
			}
		}(i, item)
	}

	// Handle quit - this one blocks
	for range quitItem.ClickedCh {
		t.logger.Info("quit clicked")
		if t.onQuit != nil {
			t.onQuit()
		}
		systray.Quit()
		return
	}
}

// toggleEnabled toggles the enabled state.
func (t *Tray) toggleEnabled() {
	t.logger.Info("toggleEnabled called", "current", t.enabled)
	t.enabled = !t.enabled

	if t.enabled {
		t.statusItem.SetTitle("✓ Enabled")
		systray.SetIcon(keyboardIcon)
	} else {
		t.statusItem.SetTitle("✗ Disabled")
		systray.SetIcon(keyboardDisabledIcon)
	}

	t.updateTooltip()

	if t.onToggle != nil {
		t.onToggle(t.enabled)
	}
}

// selectLayout changes the current layout.
func (t *Tray) selectLayout(layout string) {
	t.logger.Info("selectLayout called", "requested", layout, "current", t.currentLayout)

	if layout == t.currentLayout {
		t.logger.Info("layout already selected, skipping")
		return
	}

	// Update menu checkmarks
	for i, l := range t.availableLayouts {
		if l == layout {
			t.logger.Debug("checking layout", "layout", l, "index", i)
			t.layoutItems[i].Check()
		} else {
			t.layoutItems[i].Uncheck()
		}
	}

	t.currentLayout = layout
	t.layoutMenu.SetTitle(layout + "    ")
	t.updateTooltip()
	t.logger.Info("layout changed", "layout", layout)

	if t.onLayoutChange != nil {
		t.onLayoutChange(layout)
	}
}

func (t *Tray) updateTooltip() {
	status := "Enabled"
	if !t.enabled {
		status = "Disabled"
	}
	systray.SetTooltip("Asahi-Map: " + status + " (" + t.currentLayout + ")")
}

func (t *Tray) onExit() {
	t.logger.Info("tray exiting")
}

func (t *Tray) Quit() {
	systray.Quit()
}

func (t *Tray) SetEnabled(enabled bool) {
	t.enabled = enabled
	if t.statusItem != nil {
		if enabled {
			t.statusItem.SetTitle("✓ Enabled")
		} else {
			t.statusItem.SetTitle("✗ Disabled")
		}
	}
	t.updateTooltip()
}
