package updatercopier

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ExtractHelper(extractPath string) (string, error) {
	var helperNameBuilder strings.Builder

	switch runtime.GOOS {
	case "darwin":
		helperNameBuilder.WriteString("assets/update-helper-darwin")
	default:
		helperNameBuilder.WriteString("assets/update-helper-linux")
	}

	switch runtime.GOARCH {
	case "amd64":
		helperNameBuilder.WriteString("-amd64")
	case "arm64":
		helperNameBuilder.WriteString("-arm64")
	}

	data, err := embeddedHelpers.ReadFile(helperNameBuilder.String())
	if err != nil {
		return "", fmt.Errorf("failed to read embedded helper: %w", err)
	}

	tmpPath := filepath.Join(extractPath, "/gophorth-helper")

	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return "", fmt.Errorf("failed to write helper: %w", err)
	}

	return tmpPath, nil
}
