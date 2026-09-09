package archive

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func addFilesToTar(ctx context.Context, tw *tar.Writer, opts *CompressOptions) error {
	totalSize := int64(0)

	for _, path := range opts.FileList {
		err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if opts.SkipSymlinks && fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			if !filepath.IsAbs(file) && !filepath.IsLocal(file) {
				return fmt.Errorf("unsafe path: %s", file)
			}

			if len(opts.IncludePatterns) > 0 && !matchesAny(file, opts.IncludePatterns) {
				return nil
			}
			if len(opts.ExcludePatterns) > 0 && matchesAny(file, opts.ExcludePatterns) {
				return nil
			}

			if fi.Mode().IsRegular() && fi.Size() > opts.MaxFileSize {
				return fmt.Errorf("file too large (%s): %d bytes", file, fi.Size())
			}

			totalSize += fi.Size()
			if totalSize > opts.MaxTotalSize {
				return fmt.Errorf("archive exceeded total size limit: %d > %d", totalSize, opts.MaxTotalSize)
			}

			if opts.OnFile != nil {
				if err := opts.OnFile(file, fi.Size()); err != nil {
					return err
				}
			}

			rel, err := filepath.Rel(filepath.Dir(path), file)
			if err != nil {
				return err
			}
			if rel == "." {
				return nil
			}

			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}
			hdr.Name = rel
			if !opts.PreservePermissions {
				hdr.Mode = 0644
			}

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			if fi.Mode().IsRegular() {
				src, err := os.Open(file)
				if err != nil {
					return err
				}
				defer src.Close()

				written, err := io.Copy(tw, io.LimitReader(src, opts.MaxFileSize))
				if err != nil {
					return err
				}

				totalSize += written
				if totalSize > opts.MaxTotalSize {
					return errors.New("archive exceeded total size limit")
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
