package archive

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	kzlib "github.com/klauspost/compress/gzip"
	kzstd "github.com/klauspost/compress/zstd"
)

func extractTarRaw(ctx context.Context, src, dest string, opts *ExtractOptions) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return untarStream(ctx, f, dest, opts)
}

func extractTarGzip(ctx context.Context, src, dest string, opts *ExtractOptions) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := kzlib.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	return untarStream(ctx, gr, dest, opts)
}

func extractTarZstd(ctx context.Context, src, dest string, opts *ExtractOptions) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	zr, err := kzstd.NewReader(f, kzstd.WithDecoderMaxMemory(uint64(opts.MaxTotalSize)))
	if err != nil {
		return fmt.Errorf("failed to create zstd reader: %v", err)
	}
	defer zr.Close()

	return untarStream(ctx, zr, dest, opts)
}

func matchesPattern(name string, includes, excludes []string) bool {
	for _, ex := range excludes {
		if ok, _ := filepath.Match(ex, name); ok {
			return false
		}
	}
	if len(includes) == 0 {
		return true
	}
	for _, in := range includes {
		if ok, _ := filepath.Match(in, name); ok {
			return true
		}
	}
	return false
}

func untarStream(ctx context.Context, r io.Reader, dest string, opts *ExtractOptions) error {
	tr := tar.NewReader(r)
	var total int64
	var extractedFiles []string
	var canceled bool

	defer func() {
		if opts.OnCancelCleanup && canceled && len(extractedFiles) > 0 {
			if err := cleanupExtraction(dest, extractedFiles); err != nil {
				if opts.Logger != nil {
					opts.Logger.Printf("Cleanup warning: %v", err)
				}
			}
		}
		// Clean up any empty directories that might remain
		if err := cleanupEmptyDirectories(dest); err != nil {
			if opts.Logger != nil {
				opts.Logger.Printf("Cleanup warning: %v", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			canceled = true
			return ctx.Err()
		default:
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(hdr.Name, "./")
		if !matchesPattern(name, opts.IncludePatterns, opts.ExcludePatterns) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			target, err := verifiedPath(dest, hdr.Name)
			if err != nil {
				return err
			}
			perm := os.FileMode(hdr.Mode)
			if !opts.PreservePermissions {
				perm = 0755
			}
			if err := os.MkdirAll(target, perm); err != nil {
				return err
			}
			extractedFiles = append(extractedFiles, target)

		case tar.TypeReg:
			targetPath, err := verifiedPath(dest, hdr.Name)
			if err != nil {
				return err
			}
			extractedFiles = append(extractedFiles, targetPath)

			if _, statErr := os.Stat(filepath.Dir(targetPath)); statErr != nil {
				if errors.Is(statErr, os.ErrNotExist) {
					if basePathCreateErr := os.MkdirAll(filepath.Dir(targetPath), 0755); basePathCreateErr != nil {
						return fmt.Errorf("failed to create parent directory for symlink %s: %w", hdr.Name, basePathCreateErr)
					}
				}
			}

			err = safeWriteFile(ctx, tr, hdr.Name, fs.FileMode(hdr.Mode), dest, opts, &total)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					canceled = true
				}
				return err
			}

		case tar.TypeSymlink:

			targetPath, verifyPathErr := verifiedPath(dest, hdr.Name)
			if verifyPathErr != nil {
				return verifyPathErr
			}
			// Even if we are skipping symlinks, make sure that the path is available
			// There are cases (such as python) where an error will be thrown if not yet created
			if _, statErr := os.Stat(filepath.Dir(targetPath)); statErr != nil {
				if errors.Is(statErr, os.ErrNotExist) {
					if basePathCreateErr := os.MkdirAll(filepath.Dir(targetPath), 0755); basePathCreateErr != nil {
						return fmt.Errorf("failed to create parent directory for symlink %s: %w", hdr.Name, basePathCreateErr)
					}
				}
			}

			if !opts.SkipSymlinks {
				continue
			}

			// Security check: ensure link target doesn’t escape destination
			linkTarget := hdr.Linkname
			if filepath.IsAbs(linkTarget) {
				if opts.Logger != nil {
					opts.Logger.Printf("Warning: skipping absolute symlink %s -> %s", hdr.Name, hdr.Linkname)
				}
				continue
			}

			// Prevent path traversal (e.g. linkTarget = ../../etc/passwd)
			resolvedTarget := filepath.Join(filepath.Dir(targetPath), linkTarget)
			cleanResolved := filepath.Clean(resolvedTarget)
			if !strings.HasPrefix(cleanResolved, dest+string(os.PathSeparator)) && cleanResolved != dest {
				return fmt.Errorf("symlink %s points outside extraction root: %s", hdr.Name, hdr.Linkname)
			}

			// Remove any existing file before creating the symlink
			if removeErr := os.RemoveAll(targetPath); removeErr != nil && !os.IsNotExist(removeErr) {
				return fmt.Errorf("failed to remove existing path for symlink %s: %w", targetPath, removeErr)
			}

			if symlinkErr := os.Symlink(linkTarget, targetPath); symlinkErr != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", hdr.Name, hdr.Linkname, symlinkErr)
			}
			extractedFiles = append(extractedFiles, targetPath)

		default:
			if opts.Logger != nil {
				opts.Logger.Printf("Skipping unsupported tar entry: %s (%c)", hdr.Name, hdr.Typeflag)
			}
		}
	}
}
