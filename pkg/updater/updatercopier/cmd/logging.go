package main

import (
	"os"
	"path/filepath"
	"runtime"
)

func openLogFile(requested string) (*os.File, string) {
	// 1) If user provided a path, try it.
	if requested != "" {
		if f, err := tryCreateLog(requested); err == nil {
			return f, requested
		}
	}

	// 2) Platform-appropriate default.
	defaultPath := defaultLogPath()
	if f, err := tryCreateLog(defaultPath); err == nil {
		return f, defaultPath
	}

	// 3) Fallback to temp dir.
	fallback := filepath.Join(os.TempDir(), defaultLogFileName)
	f, _ := tryCreateLog(fallback)
	return f, fallback
}

func tryCreateLog(p string) (*os.File, error) {
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	return os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

func defaultLogPath() string {
	switch runtime.GOOS {
	case "windows":
		// Prefer LOCALAPPDATA\<App>\Logs\<file>
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, appName, "Logs", defaultLogFileName)
		}
		return filepath.Join(os.TempDir(), defaultLogFileName)

	case "darwin":
		home, _ := os.UserHomeDir()
		if home == "" {
			return filepath.Join(os.TempDir(), defaultLogFileName)
		}
		return filepath.Join(home, "Library", "Logs", appName, defaultLogFileName)

	default:
		// Linux and others:
		// Prefer XDG_STATE_HOME, then XDG_CACHE_HOME, then ~/.local/state
		if base := os.Getenv("XDG_STATE_HOME"); base != "" {
			return filepath.Join(base, appName, defaultLogFileName)
		}
		if base := os.Getenv("XDG_CACHE_HOME"); base != "" {
			return filepath.Join(base, appName, defaultLogFileName)
		}
		home, _ := os.UserHomeDir()
		if home == "" {
			return filepath.Join(os.TempDir(), defaultLogFileName)
		}
		return filepath.Join(home, ".local", "state", appName, defaultLogFileName)
	}
}
