//go:build darwin
// +build darwin

package watcher

import (
	"fmt"
	"os/exec"
	"strings"
)

func getActiveWindowMacOS() (WindowInfo, error) {
	// Use AppleScript to get active window
	script := `tell application "System Events"
		set frontApp to first application process whose frontmost is true
		set appName to name of frontApp
		try
			set windowTitle to name of front window of frontApp
			return appName & "|" & windowTitle
		on error
			return appName & "|"
		end try
	end tell`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return WindowInfo{}, err
	}

	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
	if len(parts) >= 2 {
		return WindowInfo{App: parts[0], Title: parts[1]}, nil
	}
	if len(parts) == 1 {
		return WindowInfo{App: parts[0], Title: ""}, nil
	}

	return WindowInfo{}, fmt.Errorf("could not parse window info")
}

// Stubs for other platforms (not compiled on macOS)
func getActiveWindowLinux() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("Linux not supported")
}

func getActiveWindowWindows() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("Windows not supported")
}
