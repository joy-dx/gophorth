//go:build !windows

package main

import (
	"os"
	"time"
)

func scheduleSelfDelete(logFile *os.File) {
	self := os.Args[0]

	go func() {
		timeSleep(1 * time.Second)
		// On mac/linux unlinking self is generally fine.
		if err := os.Remove(self); err != nil {
			logLine(logFile, "Error removing helper (%s): %v", self, err)
		} else {
			logLine(logFile, "Helper self-deleted successfully.")
		}
	}()
}
