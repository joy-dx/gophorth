package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CompressOptions holds configuration settings for archive Compression.
type CompressOptions struct {
	Destination         string
	MaxTotalSize        int64
	MaxFileSize         int64
	SkipSymlinks        bool
	PreservePermissions bool
	IncludePatterns     []string
	ExcludePatterns     []string
	// Path of individual files / directories to recursively go down
	FileList        []string
	OnFile          func(path string, size int64) error
	OnCancelCleanup bool
}

// DefaultCompressOptions returns safe default Compression settings.
func DefaultCompressOptions() *CompressOptions {
	return &CompressOptions{
		MaxTotalSize: 5 << 30,   // 5 GB total limit
		MaxFileSize:  500 << 20, // 500 MB/file
	}
}

// Compress dispatches to the appropriate Compressor based on the file type.
func Compress(ctx context.Context, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	if err := os.MkdirAll(filepath.Dir(opts.Destination), 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(opts.Destination))
	switch {
	case strings.HasSuffix(opts.Destination, ".tar.gz"), ext == ".tgz", ext == ".gz":
		return CompressTarGzip(ctx, opts)
	case strings.HasSuffix(opts.Destination, ".tar.zst"), ext == ".zst":
		return CompressTarZstd(ctx, opts)
	case ext == ".tar":
		return CompressTarRaw(ctx, opts)
	case ext == ".zip":
		return CompressZip(ctx, opts)
	default:
		return fmt.Errorf("unsupported archive type: %s", opts.Destination)
	}
}
