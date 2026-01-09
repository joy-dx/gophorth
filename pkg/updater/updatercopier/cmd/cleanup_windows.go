//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func scheduleSelfDelete(logFile *os.File) {
	self := os.Args[0]
	absSelf, err := filepath.Abs(self)
	if err == nil {
		self = absSelf
	}

	escaped := escapeForCmdLiteral(self)

	// Delay then delete.
	// We keep /V:OFF to avoid ! expansion surprises.
	// We use /S /C so cmd's quoting behavior is more consistent.
	command := "ping 127.0.0.1 -n 3 >NUL & del /F /Q \"" + escaped + "\""

	cmd := exec.Command("cmd.exe", "/V:OFF", "/S", "/C", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS,
	}

	if err := cmd.Start(); err != nil {
		logLine(logFile, "Failed to schedule self-delete via cmd.exe: %v", err)
		return
	}

	logLine(logFile, "Scheduled self-delete for helper: %s", self)
}

// escapeForCmdLiteral escapes characters that cmd.exe treats specially, so the
// path is safe inside a quoted string.
// We escape: ^ & | < > ( ) ! and "
//
// Note: Since we wrap the final path in double-quotes, we only need to escape
// embedded quotes and meta chars.
func escapeForCmdLiteral(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch r {
		case '^', '&', '|', '<', '>', '(', ')', '!', '"':
			out = append(out, '^', r)
		default:
			out = append(out, r)
		}
	}
	return string(out)
}
