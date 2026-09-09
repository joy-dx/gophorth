package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// used as a fallback to create log path if no path given
	appName             = "gophorth"
	defaultLogFileName  = "gophorth-update.log"
	replaceAttempts     = 15
	replaceAttemptDelay = 3 * time.Second
)

// aliases for functions to make testing simpler
var (
	osRename    = os.Rename
	osRemoveAll = os.RemoveAll
	timeSleep   = time.Sleep
)

func main() {
	pathToReplace, replacementPath, requestedLogPath, launchArgs, err :=
		parseArgs(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		_, _ = fmt.Fprintln(os.Stderr,
			"Usage: update-helper <old_path> <new_path> [log_path] [-- <args...>] [--args \"...\"]",
		)
		os.Exit(1)
	}

	pathToReplace = normalizePath(pathToReplace)
	replacementPath = normalizePath(replacementPath)

	logFile, logPath := openLogFile(requestedLogPath)
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	logLine(logFile, "Writing log to: %s", logPath)
	logLine(logFile, "Updater starting. old=%s new=%s", pathToReplace, replacementPath)
	logLine(logFile, "Launch args: %q", launchArgs)

	// Windows convention enforcement: ensure .exe when replacing a file target.
	if runtime.GOOS == "windows" && filepath.Ext(pathToReplace) == "" {
		pathToReplace += ".exe"
		logLine(logFile, "Windows: normalized target to %s", pathToReplace)
	}

	backupPath := pathToReplace + ".bak"

	logLine(logFile, "Creating backup at %s", backupPath)
	if err := copyPath(pathToReplace, backupPath); err != nil {
		logLine(logFile, "Backup failed: %v", err)
		os.Exit(1)
	}

	logLine(logFile, "Starting replace process (attempts=%d, delay=%s)",
		replaceAttempts, replaceAttemptDelay)

	if err := replaceWithRetry(logFile, pathToReplace, replacementPath); err != nil {
		logLine(logFile, "Replacement failed: %v", err)
		logLine(logFile, "Restoring backup.")
		restoreBackup(logFile, backupPath, pathToReplace, launchArgs)
		os.Exit(2)
	}

	logLine(logFile, "Attempting to launch new target")
	if err := launchApp(logFile, pathToReplace, launchArgs); err != nil {
		logLine(logFile, "Launch failed: %v", err)
		logLine(logFile, "Rolling back to backup.")
		restoreBackup(logFile, backupPath, pathToReplace, launchArgs)
		os.Exit(3)
	}

	logLine(logFile, "New target launched successfully.")
	cleanupBackup(logFile, backupPath)
	scheduleSelfDelete(logFile)
	logLine(logFile, "Helper finished.")
}

func normalizePath(p string) string {
	p = filepath.Clean(p)
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

func replaceWithRetry(logFile *os.File, targetPath, replacementPath string) error {
	var lastErr error

	for i := 0; i < replaceAttempts; i++ {
		if err := removeTarget(targetPath); err != nil {
			lastErr = err
			logLine(logFile, "Failed to remove target (attempt %d/%d): %v",
				i+1, replaceAttempts, err)
			timeSleep(replaceAttemptDelay)
			continue
		}

		// Attempt rename first (fast/atomic when possible).
		if err := osRename(replacementPath, targetPath); err == nil {
			logLine(logFile, "Replaced using rename.")
			return nil
		} else {
			lastErr = err
			logLine(logFile, "Rename failed (attempt %d/%d): %v",
				i+1, replaceAttempts, err)
		}

		// Fallback: copy+remove (handles cross-volume / different drive).
		if err := copyPath(replacementPath, targetPath); err != nil {
			lastErr = err
			logLine(logFile, "Copy fallback failed (attempt %d/%d): %v",
				i+1, replaceAttempts, err)
			timeSleep(replaceAttemptDelay)
			continue
		}
		if err := osRemoveAll(replacementPath); err != nil {
			// Not fatal; we successfully replaced target.
			logLine(logFile, "Warning: failed to remove replacement source %s: %v",
				replacementPath, err)
		}

		logLine(logFile, "Replaced using copy+remove fallback.")
		return nil
	}

	if lastErr == nil {
		lastErr = errors.New("replacement failed for unknown reasons")
	}
	return fmt.Errorf("replacement failed after %d attempts: %w", replaceAttempts, lastErr)
}

func removeTarget(targetPath string) error {
	_, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Only macOS uses RemoveAll in case we are dealing with .app
	if runtime.GOOS == "darwin" {
		return osRemoveAll(targetPath)
	}

	return os.Remove(targetPath)
}

func cleanupBackup(logFile *os.File, backupPath string) {
	logLine(logFile, "Cleaning backup at %s", backupPath)
	if err := osRemoveAll(backupPath); err != nil {
		logLine(logFile, "Error cleaning backup %s: %v", backupPath, err)
	}
}

// Detects if the path is a directory (.app on mac) and copies accordingly.
func copyPath(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// If src is a symlink, preserve it (on Unix); on Windows symlinks may require privileges.
	if info.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	_ = clearReadOnly(dst)

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Ensure destination directory exists.
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	if err := dstFile.Sync(); err != nil {
		return err
	}

	// Preserve mode on Unix-like systems.
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create dst directory.
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

		if err := copyEntry(srcPath, dstPath, entry); err != nil {
			return err
		}
	}

	// Preserve mode on Unix-like systems for directories as well (best-effort).
	if runtime.GOOS != "windows" {
		_ = os.Chmod(dst, info.Mode())
	}

	return nil
}

func copyEntry(srcPath, dstPath string, entry os.DirEntry) error {
	// Use Lstat to detect symlinks correctly.
	info, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return copySymlink(srcPath, dstPath)
	}

	if entry.IsDir() {
		return copyDir(srcPath, dstPath)
	}
	return copyFile(srcPath, dstPath)
}

func copySymlink(src, dst string) error {
	// On Unix/mac this preserves symlinks; on Windows it may fail without privileges.
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}

	// Ensure dst parent exists.
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// If dst exists, remove it first (best-effort).
	_ = os.Remove(dst)

	return os.Symlink(target, dst)
}

func logLine(logFile *os.File, msg string, args ...interface{}) {
	line := fmt.Sprintf(
		"%s: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		fmt.Sprintf(msg, args...),
	)

	log.Println(line)
	if logFile == nil {
		return
	}
	if _, err := logFile.WriteString(line + "\n"); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}
