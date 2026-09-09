package archive

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		return c.r.Read(p)
	}
}

var ErrIllegalPath = func(name string) error {
	return fmt.Errorf("illegal path traversal detected: %s", name)
}

// verifiedPath ensures the resolved path is contained within destination.
func verifiedPath(dest, name string) (string, error) {
	// Resolve destination first
	destAbs, err := filepath.Abs(filepath.Clean(dest))
	if err != nil {
		return "", fmt.Errorf("invalid destination path: %w", err)
	}

	// Clean and resolve the target path
	target := filepath.Clean(filepath.Join(dest, name))
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", ErrIllegalPath(name)
	}

	// Check if target is within destination
	destSep := len(destAbs)
	if !strings.HasPrefix(targetAbs, destAbs) ||
		(len(targetAbs) > destSep && targetAbs[destSep] != filepath.Separator) {
		return "", ErrIllegalPath(name)
	}

	return target, nil
}

// safeWriteFile writes a single file securely with size, context, and quota checks.
func safeWriteFile(
	ctx context.Context,
	reader io.Reader,
	name string,
	mode fs.FileMode,
	dest string,
	opts *ExtractOptions,
	total *int64,
) error {
	path, err := verifiedPath(dest, name)
	if err != nil {
		return err
	}

	if opts.MaxTotalSize > 0 && *total > opts.MaxTotalSize {
		return fmt.Errorf("extraction limit of %d bytes exceeded", opts.MaxTotalSize)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	perm := mode
	if !opts.PreservePermissions {
		perm = 0644
	}

	flags := os.O_CREATE | os.O_WRONLY
	if opts.Overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	f, err := os.OpenFile(path, flags, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	cr := &ctxReader{ctx: ctx, r: reader}
	buf := make([]byte, 32*1024)
	var written int64

	for {
		select {
		case <-ctx.Done():
			_ = f.Close()
			_ = os.Remove(path)
			return ctx.Err()
		default:
		}

		n, readErr := cr.Read(buf)
		if n > 0 {
			w, wErr := f.Write(buf[:n])
			written += int64(w)
			if wErr != nil {
				_ = f.Close()
				_ = os.Remove(path)
				return wErr
			}
			if opts.MaxFileSize > 0 && written > opts.MaxFileSize {
				_ = f.Close()
				_ = os.Remove(path)
				return fmt.Errorf("exceeded max file size (%d bytes)", opts.MaxFileSize)
			}
			if opts.MaxTotalSize > 0 && *total+written > opts.MaxTotalSize {
				_ = f.Close()
				_ = os.Remove(path)
				return fmt.Errorf("exceeded total size limit (%d bytes)", *total+written)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = f.Close()
			_ = os.Remove(path)
			return readErr
		}
	}

	*total += written

	if opts.OnFile != nil {
		if cbErr := opts.OnFile(path, written); cbErr != nil {
			return cbErr
		}
	}

	if opts.Logger != nil {
		opts.Logger.Printf("Extracted %s (%d bytes)", name, written)
	}
	return nil
}

func matchesAny(path string, patterns []string) bool {
	for _, p := range patterns {
		ok, err := filepath.Match(p, filepath.Base(path))
		if err == nil && ok {
			return true
		}
	}
	return false
}
