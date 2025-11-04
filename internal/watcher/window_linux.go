//go:build linux
// +build linux

package watcher

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getActiveWindowLinux tries multiple methods to detect active window
// Tries methods in order from most common to least common
func getActiveWindowLinux() (WindowInfo, error) {
	// Try xdotool first (most reliable and common on desktop Linux)
	if info, err := tryXdotool(); err == nil {
		return info, nil
	}

	// Try xprop (usually pre-installed on X11 systems)
	if info, err := tryXprop(); err == nil {
		return info, nil
	}

	// Try wmctrl (alternative tool)
	if info, err := tryWmctrl(); err == nil {
		return info, nil
	}

	// Try gdbus for GNOME (often pre-installed on GNOME desktops)
	if info, err := tryGdbus(); err == nil {
		return info, nil
	}

	// Try qdbus for KDE (KDE Plasma)
	if info, err := tryQdbus(); err == nil {
		return info, nil
	}

	return WindowInfo{}, fmt.Errorf("no window detection method available. Please install one of: xdotool, xprop, wmctrl")
}

func tryXdotool() (WindowInfo, error) {
	// Get active window ID
	windowID, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return WindowInfo{}, err
	}

	windowIDStr := strings.TrimSpace(string(windowID))

	// Get window title
	title := "unknown"
	if nameOut, err := exec.Command("xdotool", "getwindowname", windowIDStr).Output(); err == nil {
		title = strings.TrimSpace(string(nameOut))
	}

	// Get window class (application name)
	app := "unknown"
	if classOut, err := exec.Command("xdotool", "getwindowclassname", windowIDStr).Output(); err == nil {
		app = strings.TrimSpace(string(classOut))
	}

	return WindowInfo{App: app, Title: title}, nil
}

func tryXprop() (WindowInfo, error) {
	// Get active window ID
	windowIDOut, err := exec.Command("xprop", "-root", "_NET_ACTIVE_WINDOW").Output()
	if err != nil {
		return WindowInfo{}, err
	}

	// Parse window ID from output like: "_NET_ACTIVE_WINDOW(WINDOW): window id # 0x3a00007"
	fields := strings.Fields(string(windowIDOut))
	if len(fields) < 5 {
		return WindowInfo{}, fmt.Errorf("invalid xprop output")
	}

	windowID := fields[len(fields)-1]

	// Get window name
	title := "unknown"
	if nameOut, err := exec.Command("xprop", "-id", windowID, "WM_NAME").Output(); err == nil {
		title = extractXpropValue(string(nameOut))
	}

	// Get window class
	app := "unknown"
	if classOut, err := exec.Command("xprop", "-id", windowID, "WM_CLASS").Output(); err == nil {
		app = extractXpropClass(string(classOut))
	}

	return WindowInfo{App: app, Title: title}, nil
}

func tryGdbus() (WindowInfo, error) {
	// Try to get window info via GNOME Shell D-Bus interface
	// This works on GNOME desktops without extra tools
	script := `
gdbus call --session \
  --dest org.gnome.Shell \
  --object-path /org/gnome/Shell \
  --method org.gnome.Shell.Eval \
  "global.display.focus_window.get_wm_class() + '|' + global.display.focus_window.get_title()" 2>/dev/null
`

	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		return WindowInfo{}, err
	}

	// Parse output like: (true, '"Code|file.go - workspace"')
	output := string(out)
	if !strings.Contains(output, "true") {
		return WindowInfo{}, fmt.Errorf("gdbus call failed")
	}

	// Extract the quoted string
	start := strings.Index(output, `"`)
	end := strings.LastIndex(output, `"`)
	if start == -1 || end == -1 || start >= end {
		return WindowInfo{}, fmt.Errorf("could not parse gdbus output")
	}

	result := output[start+1 : end]
	parts := strings.SplitN(result, "|", 2)

	if len(parts) >= 2 {
		return WindowInfo{App: parts[0], Title: parts[1]}, nil
	}

	return WindowInfo{}, fmt.Errorf("invalid gdbus response")
}

func tryWmctrl() (WindowInfo, error) {
	// wmctrl -l -x shows windows with class names
	out, err := exec.Command("wmctrl", "-l", "-x").Output()
	if err != nil {
		return WindowInfo{}, err
	}

	// Get active window ID first
	activeOut, err := exec.Command("xprop", "-root", "_NET_ACTIVE_WINDOW").Output()
	if err != nil {
		return WindowInfo{}, err
	}

	fields := strings.Fields(string(activeOut))
	if len(fields) < 5 {
		return WindowInfo{}, fmt.Errorf("invalid xprop output")
	}

	activeID := fields[len(fields)-1]

	// Parse wmctrl output to find the active window
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, activeID) {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				app := parts[2] // WM_CLASS
				title := strings.Join(parts[3:], " ")
				if idx := strings.LastIndex(app, "."); idx >= 0 {
					app = app[idx+1:]
				}
				return WindowInfo{App: app, Title: title}, nil
			}
		}
	}

	return WindowInfo{}, fmt.Errorf("could not find active window in wmctrl output")
}

func tryQdbus() (WindowInfo, error) {
	// Try KDE Plasma via qdbus
	script := `
qdbus org.kde.ActivityManager /ActivityManager/Resources CurrentActivity 2>/dev/null || \
qdbus org.kde.plasmashell /PlasmaShell org.kde.PlasmaShell.evaluateScript \
  'var win = workspace.activeClient; win ? win.resourceClass + "|" + win.caption : ""' 2>/dev/null
`

	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		return WindowInfo{}, err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return WindowInfo{}, fmt.Errorf("no output from qdbus")
	}

	parts := strings.SplitN(output, "|", 2)
	if len(parts) >= 2 {
		return WindowInfo{App: parts[0], Title: parts[1]}, nil
	}

	return WindowInfo{}, fmt.Errorf("could not parse qdbus output")
}

func extractXpropValue(output string) string {
	// Extract value from: WM_NAME(STRING) = "value"
	if idx := strings.Index(output, "= "); idx >= 0 {
		value := output[idx+2:]
		value = strings.Trim(value, "\" \n\r")
		return value
	}
	return "unknown"
}

func extractXpropClass(output string) string {
	// Extract class from: WM_CLASS(STRING) = "instance", "class"
	if idx := strings.Index(output, "= "); idx >= 0 {
		value := output[idx+2:]
		parts := strings.Split(value, ",")
		if len(parts) >= 2 {
			class := strings.Trim(parts[1], "\" \n\r")
			return class
		}
		if len(parts) == 1 {
			return strings.Trim(parts[0], "\" \n\r")
		}
	}
	return "unknown"
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
} // Stubs for other platforms (not compiled on Linux)
func getActiveWindowMacOS() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("macOS not supported")
}

func getActiveWindowWindows() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("Windows not supported")
}
