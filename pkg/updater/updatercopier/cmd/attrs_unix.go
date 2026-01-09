//go:build !windows

package main

// clearReadOnly This function is really for Windows. On posix systems, this is a stub
func clearReadOnly(path string) error {
	return nil
}
