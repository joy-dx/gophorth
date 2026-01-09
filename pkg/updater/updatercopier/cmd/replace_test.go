package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReplaceWithRetry_RenameFails_CopyFallbackWorks_Golden(t *testing.T) {
	t.Parallel()

	origRename := osRename
	origSleep := timeSleep
	defer func() {
		osRename = origRename
		timeSleep = origSleep
	}()

	osRename = func(oldpath, newpath string) error {
		return errors.New("simulated rename failure")
	}

	timeSleep = func(d time.Duration) {
		// no-op
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "app.bin")
	repl := filepath.Join(dir, "app.new")

	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(repl, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceWithRetry(nil, target, repl); err != nil {
		t.Fatalf("replaceWithRetry: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("target content got=%q want=%q", string(got), "new")
	}
}
