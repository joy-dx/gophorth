//go:build darwin

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func launchApp(logFile *os.File, path string, args []string) error {
	var cmd *exec.Cmd

	if filepath.Ext(path) == ".app" {
		logLine(logFile, ".app on darwin detected, using open -n")
		cmd = exec.Command("open", append([]string{"-n", path, "--args"}, args...)...)
	} else {
		cmd = exec.Command(path, args...)
	}

	if err := cmd.Start(); err != nil {
		logLine(logFile, "Launch failed for %s: %v", path, err)
		return err
	}

	logLine(logFile, "Launch successful: %s", path)
	return nil
}
