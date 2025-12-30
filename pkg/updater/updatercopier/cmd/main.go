package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: update-helper <old_path> <new_path> <log_path>")
		os.Exit(1)
	}

	pathToReplace := os.Args[1]
	replacementFilePath := os.Args[2]

	logPath := "./gophorth-update.log"
	if len(os.Args) >= 4 {
		logPath = os.Args[3]
	}

	log.Printf("Writing log to: %s\n", logPath)
	logFile, _ := os.Create(logPath)
	defer logFile.Close()

	logLine(logFile, "Updater starting. old=%s new=%s", pathToReplace, replacementFilePath)

	backupPath := pathToReplace + ".bak"

	// Step 1: Back up existing binary
	logLine(logFile, "Creating backup at %s", backupPath)
	if err := copyPath(pathToReplace, backupPath); err != nil {
		logLine(logFile, "Backup failed: %v", err)
		os.Exit(1)
	}

	// Step 2: Wait for main app to fully exit
	successfulRename := false
	logLine(logFile, "Starting rename process")
	for i := 0; i < 20; i++ {
		if err := os.RemoveAll(pathToReplace); err != nil {
			logLine(logFile, "Failed to remove old version (attempt %d): %v", i+1, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if err := os.Rename(replacementFilePath, pathToReplace); err != nil {
			logLine(logFile, "renaming failed attempt %d: %v", i+1, err)
			time.Sleep(1 * time.Second)
		} else {
			logLine(logFile, "Successfully replaced old binary.")
			successfulRename = true
			break
		}

	}

	if !successfulRename {
		logLine(logFile, "Replacement failed after 20 attempts. Restoring backup.")
		restoreBackup(logFile, backupPath, pathToReplace)
		os.Exit(2)
	}

	logLine(logFile, "Attempting to launch new binary")
	if err := launchApp(logFile, pathToReplace); err != nil {
		logLine(logFile, "Launch failed: %v", err)
		logLine(logFile, "Rolling back to backup.")
		restoreBackup(logFile, backupPath, pathToReplace)
		os.Exit(3)
	}

	logLine(logFile, "New binary launched successfully.")
	cleanupHelper(logFile, backupPath)
	logLine(logFile, "Helper finished.")

}

// Detects if the path is a directory (.app on mac) and copies accordingly.
func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// On macOS, .app is actually a directory
	if runtime.GOOS == "darwin" && info.IsDir() && filepath.Ext(src) == ".app" {
		return copyDir(src, dst)
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

// For regular file copy
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}

// For recursive directory copy (used for .app and fallback dirs)
func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanupHelper(logFile *os.File, backupPath string) {
	logLine(logFile, "Cleaning temporary files.")

	// Remove backup — works for files or .app directories
	if err := os.RemoveAll(backupPath); err != nil {
		logLine(logFile, fmt.Sprintf(
			"Error cleaning up backup files at %s: %v", backupPath, err,
		))
	}

	// Self-delete after leaving a short delay
	self := os.Args[0]
	go func() {
		time.Sleep(1 * time.Second)
		if err := os.RemoveAll(self); err != nil {
			logLine(logFile, fmt.Sprintf(
				"Error removing helper (%s): %v", self, err,
			))
		} else {
			logLine(logFile, "Helper self-deleted successfully.")
		}
	}()
}

// launchApp starts the target application in a platform‑safe way.
// On macOS, it supports both .app bundles (via "open -n") and regular binaries.
// On other OSes, it just launches the binary directly.
func launchApp(logFile *os.File, path string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" && filepath.Ext(path) == ".app" {
		// GUI‑friendly macOS launch
		cmd = exec.Command("open", "-n", path)
	} else {
		cmd = exec.Command(path)
	}

	if err := cmd.Start(); err != nil {
		logLine(logFile, "Launch failed for %s: %v", path, err)
		return err
	}

	logLine(logFile, "Launch successful: %s", path)
	return nil
}

func logLine(logFile *os.File, msg string, args ...interface{}) {
	entry := fmt.Sprintf("%s: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		fmt.Sprintf(msg, args...))
	log.Println(entry)
	if _, err := logFile.WriteString(entry); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

func restoreBackup(logFile *os.File, backupPath, targetPath string) {
	logLine(logFile, "Restoring backup from %s to %s", backupPath, targetPath)

	// Remove existing broken version (handles .app directories)
	if err := os.RemoveAll(targetPath); err != nil {
		logLine(logFile, "Failed to remove unwanted new version: %v", err)
	}

	// Move backup into place
	if err := os.Rename(backupPath, targetPath); err != nil {
		logLine(logFile, "Failed to restore backup: %v", err)
		return
	}

	if err := launchApp(logFile, targetPath); err != nil {
		logLine(logFile, "Failed to start restored version: %v", err)
	} else {
		logLine(logFile, "Old version relaunched successfully.")
	}
}
