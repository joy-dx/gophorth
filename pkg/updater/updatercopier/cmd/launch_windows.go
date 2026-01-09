//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func launchApp(logFile *os.File, path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS,
	}

	if err := cmd.Start(); err != nil {
		logLine(logFile, "Launch failed for %s: %v", path, err)
		return err
	}

	logLine(logFile, "Launch successful: %s", path)
	return nil
}
