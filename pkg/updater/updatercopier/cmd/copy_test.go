package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyFile_PreservesContent_Golden(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	want := "hello\nworld\n"
	if err := os.WriteFile(src, []byte(want), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	gotBytes, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	if got != want {
		t.Fatalf("content mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}

	if runtime.GOOS != "windows" {
		si, _ := os.Stat(src)
		di, _ := os.Stat(dst)
		if di.Mode() != si.Mode() {
			t.Fatalf("mode mismatch got=%v want=%v", di.Mode(), si.Mode())
		}
	}
}

func TestCopyPath_SymlinkPreservedOnUnix_Golden(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink creation often requires privileges on Windows")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	dst := filepath.Join(dir, "dstlink.txt")

	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := copyPath(link, dst); err != nil {
		t.Fatalf("copyPath symlink: %v", err)
	}

	// Ensure dst is symlink and points to the same target string.
	info, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("dst is not symlink: mode=%v", info.Mode())
	}
	gotTarget, err := os.Readlink(dst)
	if err != nil {
		t.Fatal(err)
	}
	wantTarget, _ := os.Readlink(link)
	if gotTarget != wantTarget {
		t.Fatalf("symlink target mismatch got=%q want=%q", gotTarget, wantTarget)
	}
}
