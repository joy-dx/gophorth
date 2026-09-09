package main

import "os"

func restoreBackup(logFile *os.File, backupPath, targetPath string, launchArgs []string) {
	logLine(logFile, "Restoring backup from %s to %s", backupPath, targetPath)

	if err := removeTarget(targetPath); err != nil {
		logLine(logFile, "Failed to remove unwanted version: %v", err)
	}

	if err := osRename(backupPath, targetPath); err != nil {
		logLine(logFile, "Rename restore failed: %v", err)
		logLine(logFile, "Attempting restore via copy+remove fallback.")

		if err2 := copyPath(backupPath, targetPath); err2 != nil {
			logLine(logFile, "Copy restore failed: %v", err2)
			return
		}
		_ = osRemoveAll(backupPath)
	}

	if err := launchApp(logFile, targetPath, launchArgs); err != nil {
		logLine(logFile, "Failed to start restored version: %v", err)
	} else {
		logLine(logFile, "Old version relaunched successfully.")
	}
}
