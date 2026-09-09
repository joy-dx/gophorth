package archive

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func CompressZip(ctx context.Context, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	out, err := os.Create(opts.Destination)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() {
		_ = out.Close()
		if opts.OnCancelCleanup {
			select {
			case <-ctx.Done():
				_ = os.Remove(opts.Destination)
			default:
			}
		}
	}()

	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, path := range opts.FileList {
		err = filepath.Walk(path, func(file string, fi os.FileInfo, errIn error) error {
			if errIn != nil {
				return errIn
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if opts.SkipSymlinks && fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			if !filepath.IsLocal(file) {
				return fmt.Errorf("unsafe path: %s", file)
			}

			if fi.IsDir() {
				return nil // zip doesn’t require explicit folder entries
			}

			if fi.Size() > opts.MaxFileSize {
				return fmt.Errorf("file too large: %s", file)
			}

			if len(opts.IncludePatterns) > 0 && !matchesAny(file, opts.IncludePatterns) {
				return nil
			}
			if len(opts.ExcludePatterns) > 0 && matchesAny(file, opts.ExcludePatterns) {
				return nil
			}

			rel, err := filepath.Rel(filepath.Dir(path), file)
			if err != nil {
				return err
			}

			hdr, err := zip.FileInfoHeader(fi)
			if err != nil {
				return err
			}
			hdr.Name = rel
			hdr.Method = zip.Deflate
			hdr.Modified = time.Now()

			w, err := zw.CreateHeader(hdr)
			if err != nil {
				return err
			}

			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(w, io.LimitReader(f, opts.MaxFileSize))
			return err
		})
		if err != nil {
			return err
		}
	}

	return zw.Close()
}
