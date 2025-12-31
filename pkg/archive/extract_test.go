package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	kzlib "github.com/klauspost/compress/gzip"
	kzstd "github.com/klauspost/compress/zstd"
)

func writeTar(t *testing.T, files map[string]string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, content); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	return buf
}

func writeTarGzip(t *testing.T, files map[string]string) *bytes.Buffer {
	var buf bytes.Buffer
	gz := kzlib.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, content); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gz.Close()
	return &buf
}

func writeTarZstd(t *testing.T, files map[string]string) *bytes.Buffer {
	var buf bytes.Buffer
	zw, _ := kzstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, content); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	zw.Close()
	return &buf
}

func writeZip(t *testing.T, files map[string]string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}
	zw.Close()
	return buf
}

func createTestArchiveFile(t *testing.T, data *bytes.Buffer, suffix string) string {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data"+suffix)
	if err := os.WriteFile(path, data.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

type testLogger struct {
	lines []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.lines = append(l.lines, fmt.Sprintf(format, v...))
}

// ────────────────────────────────────────────────
// CORE TESTS
// ────────────────────────────────────────────────

func TestExtract_TarTypes(t *testing.T) {
	files := map[string]string{"a.txt": "aaa", "folder/b.txt": "bbb"}

	for _, tc := range []struct {
		name   string
		buffer *bytes.Buffer
		suffix string
	}{
		{"tar", writeTar(t, files), ".tar"},
		{"tgz", writeTarGzip(t, files), ".tar.gz"},
		{"zstd", writeTarZstd(t, files), ".tar.zst"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := createTestArchiveFile(t, tc.buffer, tc.suffix)
			dest := t.TempDir()
			opts := DefaultExtractOptions()
			opts.Logger = &testLogger{}

			if err := Extract(context.Background(), src, dest, opts); err != nil {
				t.Fatalf("extract failed: %v", err)
			}
			data, err := os.ReadFile(filepath.Join(dest, "a.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "aaa" {
				t.Fatalf("expected 'aaa', got %q", data)
			}
		})
	}
}

func TestExtract_Zip(t *testing.T) {
	files := map[string]string{"x.txt": "xxx", "y.txt": "yyy"}
	src := createTestArchiveFile(t, writeZip(t, files), ".zip")
	dest := t.TempDir()

	opts := DefaultExtractOptions()
	if err := Extract(context.Background(), src, dest, opts); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(dest, "x.txt"))
	if err != nil || string(out) != "xxx" {
		t.Fatalf("unexpected output: %v %s", err, out)
	}
}

func TestExtract_PreventsPathTraversal(t *testing.T) {
	files := map[string]string{"../evil.txt": "boom"}
	src := createTestArchiveFile(t, writeTar(t, files), ".tar")
	dest := t.TempDir()

	err := Extract(context.Background(), src, dest, nil)
	if err == nil || !strings.Contains(err.Error(), "illegal path") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

func TestExtract_RespectsMaxFileSize(t *testing.T) {
	files := map[string]string{"big.txt": strings.Repeat("x", 1024)}
	src := createTestArchiveFile(t, writeZip(t, files), ".zip")
	dest := t.TempDir()

	opts := DefaultExtractOptions()
	opts.MaxFileSize = 100 // very small

	err := Extract(context.Background(), src, dest, opts)
	if err == nil {
		t.Fatal("expected max file size error, got nil")
	}
}

func TestExtract_RespectsMaxTotalSize(t *testing.T) {
	files := map[string]string{
		"a.txt": strings.Repeat("a", 150),
		"b.txt": strings.Repeat("b", 150),
	}
	src := createTestArchiveFile(t, writeTar(t, files), ".tar")
	dest := t.TempDir()

	opts := DefaultExtractOptions()
	opts.MaxTotalSize = 200
	err := Extract(context.Background(), src, dest, opts)
	if err == nil || !strings.Contains(err.Error(), "exceeded total size") {
		t.Fatalf("expected total size limit hit, got: %v", err)
	}
}

func TestExtract_CancelContext(t *testing.T) {
	files := map[string]string{}
	for i := 0; i < 20; i++ {
		files[fmt.Sprintf("f%02d.txt", i)] = strings.Repeat("x", 500_000) // large enough
	}
	src := createTestArchiveFile(t, writeTar(t, files), ".tar")
	dest := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	opts := DefaultExtractOptions()
	opts.OnCancelCleanup = true
	opts.OnFile = func(path string, size int64) error {
		cancel()
		return nil
	}

	err := Extract(ctx, src, dest, opts)
	if err == nil {
		t.Fatal("expected context cancellation")
	}

	fileCount := 0
	filepath.Walk(dest, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			fileCount++
		}
		return nil
	})
	if fileCount != 0 {
		t.Fatalf("expected no residual files, found %d", fileCount)
	}
}

func TestExtract_IncludeExcludePatterns(t *testing.T) {
	files := map[string]string{
		"x.txt": "111",
		"y.txt": "222",
		"z.log": "333",
	}
	src := createTestArchiveFile(t, writeTar(t, files), ".tar")
	dest := t.TempDir()

	opts := DefaultExtractOptions()
	opts.IncludePatterns = []string{"*.txt"}
	opts.ExcludePatterns = []string{"y.*"}

	if err := Extract(context.Background(), src, dest, opts); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "x.txt")); err != nil {
		t.Fatal("x.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dest, "y.txt")); err == nil {
		t.Fatal("y.txt should be excluded")
	}
	if _, err := os.Stat(filepath.Join(dest, "z.log")); err == nil {
		t.Fatal("z.log should be excluded by include pattern")
	}
}

func TestExtract_SymlinkHandling(t *testing.T) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	// Create a symlink entry
	if err := tw.WriteHeader(&tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "target.txt",
		Mode:     0644,
	}); err != nil {
		t.Fatal(err)
	}
	tw.Close()

	src := createTestArchiveFile(t, buf, ".tar")
	dest := t.TempDir()

	opts := DefaultExtractOptions()
	opts.SkipSymlinks = true

	if err := Extract(context.Background(), src, dest, opts); err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	target := filepath.Join(dest, "link.txt")
	if fi, err := os.Lstat(target); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink created, got: %v, mode=%v", err, fi.Mode())
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	src := filepath.Join(t.TempDir(), "foo.rar")
	if err := os.WriteFile(src, []byte("not valid"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	err := Extract(context.Background(), src, dest, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported format error, got %v", err)
	}
}
