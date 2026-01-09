//go:build windows

package main

import "os"

func clearReadOnly(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()
	// Clear write-protection by adding owner-write bit.
	// This is a best-effort; Windows attributes are not 1:1 with chmod bits.
	return os.Chmod(path, mode|0o200)
}
