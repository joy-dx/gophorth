package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Logger defines a minimal logger interface used for progress output.
type Logger interface {
	Printf(format string, v ...interface{})
}

// ExtractOptions holds configuration settings for archive extraction.
type ExtractOptions struct {
	MaxTotalSize        int64
	MaxFileSize         int64
	SkipSymlinks        bool
	PreservePermissions bool
	Overwrite           bool
	IncludePatterns     []string
	ExcludePatterns     []string
	OnFile              func(path string, size int64) error
	OnCancelCleanup     bool
	Logger              Logger
}

// DefaultExtractOptions returns safe default extraction settings.
func DefaultExtractOptions() *ExtractOptions {
	return &ExtractOptions{
		MaxTotalSize:        5 << 30,   // 5 GB total limit
		MaxFileSize:         500 << 20, // 500 MB/file
		PreservePermissions: true,
	}
}

// Extract dispatches to the appropriate extractor based on the file type.
func Extract(ctx context.Context, src, dest string, opts *ExtractOptions) error {
	if opts == nil {
		opts = DefaultExtractOptions()
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(src))
	switch {
	case strings.HasSuffix(src, ".tar.gz"), strings.HasSuffix(src, ".tgz"), ext == ".gz":
		return extractTarGzip(ctx, src, dest, opts)
	case strings.HasSuffix(src, ".tar.zst"), ext == ".zst":
		return extractTarZstd(ctx, src, dest, opts)
	case ext == ".tar":
		return extractTarRaw(ctx, src, dest, opts)
	case ext == ".zip":
		return extractZip(ctx, src, dest, opts)
	default:
		return fmt.Errorf("unsupported archive type: %s", src)
	}
}
