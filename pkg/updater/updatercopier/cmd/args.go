package main

import (
	"errors"
	"fmt"
	"strings"
)

func parseArgs(argv []string) (
	oldPath string,
	newPath string,
	logPath string,
	launchArgs []string,
	err error,
) {
	// argv[0] is program name.
	if len(argv) < 3 {
		return "", "", "", nil, errors.New("missing required args")
	}

	// Expect at least: prog old new
	oldPath = argv[1]
	newPath = argv[2]

	// Remaining tokens start at 3.
	rest := argv[3:]

	// Optional 3rd positional: log path, but only if it doesn't look like a flag and isn't "--".
	if len(rest) > 0 &&
		rest[0] != "--" &&
		rest[0] != "--args" &&
		!strings.HasPrefix(rest[0], "-") {
		logPath = rest[0]
		rest = rest[1:]
	}

	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--":
			launchArgs = append(launchArgs, rest[i+1:]...)
			return oldPath, newPath, logPath, launchArgs, nil

		case "--args":
			if i+1 >= len(rest) {
				return "", "", "", nil, fmt.Errorf("--args requires a value")
			}
			parsed, perr := splitArgsString(rest[i+1])
			if perr != nil {
				return "", "", "", nil, fmt.Errorf("invalid --args: %w", perr)
			}
			launchArgs = append(launchArgs, parsed...)
			i++ // consume value
		default:
			return "", "", "", nil, fmt.Errorf("unknown token: %s", rest[i])
		}
	}

	return oldPath, newPath, logPath, launchArgs, nil
}

// splitArgsString parses a shell-ish argument string into []string.
// Supports double-quotes and backslash escapes inside quotes.
// This is intentionally minimal and deterministic (not a full shell parser).
func splitArgsString(s string) ([]string, error) {
	var out []string
	var cur []rune
	inQuotes := false
	escaped := false

	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}

	for _, r := range []rune(s) {
		if escaped {
			cur = append(cur, r)
			escaped = false
			continue
		}

		if inQuotes && r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			inQuotes = !inQuotes
			continue
		}

		if !inQuotes && (r == ' ' || r == '\t' || r == '\n') {
			flush()
			continue
		}

		cur = append(cur, r)
	}

	if escaped {
		return nil, fmt.Errorf("dangling escape")
	}
	if inQuotes {
		return nil, fmt.Errorf("unterminated quote")
	}

	flush()
	return out, nil
}
