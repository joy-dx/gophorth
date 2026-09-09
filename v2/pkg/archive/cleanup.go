package archive

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func cleanupExtraction(dest string, extractedFiles []string) error {
	var errs []error
	// Clean up in reverse order to handle dependencies
	for i := len(extractedFiles) - 1; i >= 0; i-- {
		path := extractedFiles[i]
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}

	// Clean up empty parent directories
	for _, path := range extractedFiles {
		dir := filepath.Dir(path)
		for dir != "" && dir != filepath.Dir(dest) {
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				if !errors.Is(err, fs.ErrInvalid) && !errors.Is(err, fs.ErrPermission) {
					errs = append(errs, fmt.Errorf("cleanup dir %s: %w", dir, err))
				}
				break
			}
			dir = filepath.Dir(dir)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

func cleanupEmptyDirectories(dest string) error {
	return filepath.WalkDir(dest, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors during cleanup
		}
		if d.IsDir() && path != dest {
			empty, err := isDirectoryEmpty(path)
			if err != nil {
				return nil
			}
			if empty {
				_ = os.Remove(path)
			}
		}
		return nil
	})
}

func isDirectoryEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Read([]byte{0})
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
