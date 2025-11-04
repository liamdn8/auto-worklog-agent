package watcher

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// WindowInfo represents the currently active window.
type WindowInfo struct {
	App   string
	Title string
}

// GetActiveWindow returns information about the currently focused window.
// Platform-specific implementations are in window_linux.go, window_darwin.go, window_windows.go
func GetActiveWindow() (WindowInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return getActiveWindowLinux()
	case "darwin":
		return getActiveWindowMacOS()
	case "windows":
		return getActiveWindowWindows()
	default:
		return WindowInfo{}, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// WatchActiveWindow polls for active window changes and sends events.
func WatchActiveWindow(pollInterval time.Duration, callback func(WindowInfo)) {
	var lastWindow WindowInfo
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		window, err := GetActiveWindow()
		if err != nil {
			log.Printf("Failed to get active window: %v", err)
			continue
		}

		// Only trigger callback if window changed
		if window.App != lastWindow.App || window.Title != lastWindow.Title {
			lastWindow = window
			callback(window)
		}
	}
}
