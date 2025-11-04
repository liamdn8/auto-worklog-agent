//go:build windows
// +build windows

package watcher

import (
	"fmt"
	"os/exec"
	"strings"
)

func getActiveWindowWindows() (WindowInfo, error) {
	// Use PowerShell to get active window
	script := `Add-Type @"
		using System;
		using System.Runtime.InteropServices;
		using System.Text;
		public class Win32 {
			[DllImport("user32.dll")]
			public static extern IntPtr GetForegroundWindow();
			[DllImport("user32.dll")]
			public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
			[DllImport("user32.dll", SetLastError=true)]
			public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
		}
"@
	$hwnd = [Win32]::GetForegroundWindow()
	$title = New-Object System.Text.StringBuilder 256
	[void][Win32]::GetWindowText($hwnd, $title, $title.Capacity)
	$processId = 0
	[void][Win32]::GetWindowThreadProcessId($hwnd, [ref]$processId)
	$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
	if ($process) {
		Write-Output "$($process.ProcessName)|$($title.ToString())"
	} else {
		Write-Output "unknown|$($title.ToString())"
	}`

	out, err := exec.Command("powershell", "-Command", script).Output()
	if err != nil {
		return WindowInfo{}, err
	}

	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
	if len(parts) >= 2 {
		return WindowInfo{App: parts[0], Title: parts[1]}, nil
	}

	return WindowInfo{}, fmt.Errorf("could not parse window info")
}

// Stubs for other platforms (not compiled on Windows)
func getActiveWindowLinux() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("Linux not supported")
}

func getActiveWindowMacOS() (WindowInfo, error) {
	return WindowInfo{}, fmt.Errorf("macOS not supported")
}
