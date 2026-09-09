//go:build windows

package main

import (
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
)

func launchApp(logFile *os.File, path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
	}

	if err := cmd.Start(); err != nil {
		logLine(logFile, "Launch failed for %s: %v", path, err)
		return err
	}

	logLine(logFile, "Launch successful: %s", path)
	return nil
}
