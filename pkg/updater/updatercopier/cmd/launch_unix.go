//go:build !windows && !darwin

package main

import (
	"os"
	"os/exec"
)

func launchApp(logFile *os.File, path string, args []string) error {
	cmd := exec.Command(path, args...)
	if err := cmd.Start(); err != nil {
		logLine(logFile, "Launch failed for %s: %v", path, err)
		return err
	}

	logLine(logFile, "Launch successful: %s", path)
	return nil
}
