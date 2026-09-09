package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultLogPath_Golden(t *testing.T) {
	t.Parallel()

	// Snapshot env vars we touch
	oldLocal := os.Getenv("LOCALAPPDATA")
	oldState := os.Getenv("XDG_STATE_HOME")
	oldCache := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		_ = os.Setenv("LOCALAPPDATA", oldLocal)
		_ = os.Setenv("XDG_STATE_HOME", oldState)
		_ = os.Setenv("XDG_CACHE_HOME", oldCache)
	}()

	switch runtime.GOOS {
	case "windows":
		_ = os.Setenv("LOCALAPPDATA", `C:\Users\me\AppData\Local`)
		got := defaultLogPath()
		want := filepath.Join(`C:\Users\me\AppData\Local`, appName, "Logs", defaultLogFileName)
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}

	case "darwin":
		// Hard to force home without OS-specific hooks; just assert suffix.
		got := defaultLogPath()
		suffix := filepath.Join("Library", "Logs", appName, defaultLogFileName)
		if !hasSuffixPath(got, suffix) {
			t.Fatalf("got %q want suffix %q", got, suffix)
		}

	default:
		_ = os.Setenv("XDG_STATE_HOME", "/tmp/state")
		_ = os.Setenv("XDG_CACHE_HOME", "/tmp/cache")
		got := defaultLogPath()
		want := filepath.Join("/tmp/state", appName, defaultLogFileName)
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func hasSuffixPath(p, suffix string) bool {
	// filepath.Clean for stable comparisons.
	p = filepath.Clean(p)
	suffix = filepath.Clean(suffix)
	if len(p) < len(suffix) {
		return false
	}
	return p[len(p)-len(suffix):] == suffix
}
