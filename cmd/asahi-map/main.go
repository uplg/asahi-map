// Asahi-Map: Lightweight macOS Option key shortcut handler for Linux
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/uplg/asahi-map/internal/config"
	"github.com/uplg/asahi-map/internal/handler"
	"github.com/uplg/asahi-map/internal/keyboard"
	"github.com/uplg/asahi-map/internal/mappings"
	"github.com/uplg/asahi-map/internal/tray"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to config file")
	layoutName := flag.String("layout", "", "Layout name to use")
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	showVersion := flag.Bool("version", false, "Show version information")
	noTray := flag.Bool("no-tray", false, "Run without system tray")
	flag.Parse()

	if *showVersion {
		fmt.Printf("asahi-map %s (%s) built %s\n", version, commit, buildDate)
		os.Exit(0)
	}

	// Setup logging
	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Override layout if specified on command line
	if *layoutName != "" {
		cfg.Layout = *layoutName
	}

	logger.Info("asahi-map starting",
		"version", version,
		"layout", cfg.Layout,
	)

	// Create config directory if needed
	if err := ensureConfigDir(cfg); err != nil {
		logger.Error("failed to create config directory", "error", err)
		os.Exit(1)
	}

	// Load layout
	layoutPath := cfg.LayoutPath(cfg.Layout)
	logger.Debug("loading layout file", "path", layoutPath)
	layout, err := mappings.LoadLayout(layoutPath)
	if err != nil {
		logger.Error("failed to load layout", "layout", cfg.Layout, "path", layoutPath, "error", err)
		os.Exit(1)
	}
	logger.Info("loaded layout", "name", layout.Name, "description", layout.Description, "path", layoutPath)

	// Create key lookup
	lookup := mappings.NewKeyLookup(layout)

	// Create virtual keyboard
	vkb, err := keyboard.NewVirtualKeyboard(logger)
	if err != nil {
		logger.Error("failed to create virtual keyboard", "error", err)
		logger.Error("make sure you have write access to /dev/uinput")
		os.Exit(1)
	}
	defer vkb.Close()

	// Find and grab keyboard devices
	devManager := keyboard.NewDeviceManager(logger)
	defer devManager.Close()

	keyboards, err := devManager.FindKeyboards()
	if err != nil {
		logger.Error("failed to find keyboards", "error", err)
		os.Exit(1)
	}

	if len(keyboards) == 0 {
		logger.Error("no keyboards found")
		os.Exit(1)
	}

	// Grab the first keyboard (or all if needed)
	for _, kb := range keyboards {
		if err := devManager.GrabDevice(kb); err != nil {
			logger.Error("failed to grab keyboard", "name", kb.Name(), "error", err)
			continue
		}
	}

	// Create event channel
	events := make(chan *keyboard.KeyEvent, 100)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start reading events from keyboards
	for _, kb := range keyboards {
		go func(dev *keyboard.Device) {
			if err := keyboard.ReadEvents(ctx, dev, events); err != nil {
				logger.Error("error reading events", "device", dev.Name(), "error", err)
			}
		}(kb)
	}

	// Create handler
	h := handler.New(lookup, vkb, logger)

	// Start event processing in background
	go func() {
		if err := h.ProcessEvents(ctx, events); err != nil {
			logger.Error("error processing events", "error", err)
		}
	}()

	// Get available layouts for tray menu
	availableLayouts, err := cfg.AvailableLayouts()
	if err != nil {
		logger.Warn("could not list layouts", "error", err)
		availableLayouts = []string{cfg.Layout}
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *noTray {
		// Run without tray, wait for signal
		logger.Info("running without system tray, press Ctrl+C to quit")
		<-sigChan
		logger.Info("shutting down...")
	} else {
		// Create and run system tray
		trayCfg := tray.Config{
			CurrentLayout:    cfg.Layout,
			AvailableLayouts: availableLayouts,
			Enabled:          true,
			OnLayoutChange: func(layoutName string) {
				newLayout, err := mappings.LoadLayout(cfg.LayoutPath(layoutName))
				if err != nil {
					logger.Error("failed to load layout", "layout", layoutName, "error", err)
					return
				}
				cfg.Layout = layoutName
				cfg.Save()
				h.SetLayout(mappings.NewKeyLookup(newLayout))
			},
			OnToggle: func(enabled bool) {
				h.SetEnabled(enabled)
			},
			OnQuit: func() {
				logger.Info("shutting down...")
				cancel()
				os.Exit(0)
			},
			Logger: logger,
		}

		trayIcon := tray.New(trayCfg)

		// Handle signals in a goroutine
		go func() {
			<-sigChan
			logger.Info("shutting down...")
			trayIcon.Quit()
		}()

		// Run systray (blocks)
		trayIcon.Run()
	}

	logger.Info("asahi-map stopped")
}

// ensureConfigDir creates the config directory and copies default layouts if needed.
func ensureConfigDir(cfg *config.Config) error {
	layoutDir := filepath.Join(cfg.ConfigDir, "layouts")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		return err
	}
	return nil
}
