package archive

import (
	"context"
	"errors"
	"os"

	kzip "github.com/klauspost/compress/zip"
)

func extractZip(ctx context.Context, src, dest string, opts *ExtractOptions) error {
	r, err := kzip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

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

	for _, f := range r.File {
		select {
		case <-ctx.Done():
			canceled = true
			return ctx.Err()
		default:
		}

		if !matchesPattern(f.Name, opts.IncludePatterns, opts.ExcludePatterns) {
			continue
		}

		path, err := verifiedPath(dest, f.Name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			extractedFiles = append(extractedFiles, path)
			continue
		}

		if opts.Logger != nil {
			opts.Logger.Printf("Extracting %s", f.Name)
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		extractedFiles = append(extractedFiles, path)
		err = safeWriteFile(ctx, rc, f.Name, f.Mode(), dest, opts, &total)
		rc.Close()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				canceled = true
			}
			return err
		}
	}
	return nil
}
