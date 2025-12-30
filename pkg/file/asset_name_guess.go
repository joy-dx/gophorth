package file

import (
	"path/filepath"
	"strings"
)

// AssetNameGuess guesses platform and arch from asset names.
// It recognizes common patterns across OS/arch and archives.
//
// Examples of matches:
//   - mytool_1.2.3_linux_amd64.tar.gz -> linux, amd64
//   - mytool-windows-amd64.zip -> windows, amd64
//   - mytool_darwin_arm64.zip -> darwin, arm64
//   - mytool_linux_x86_64_musl.tar.gz -> linux, amd64
//   - mytool_linux_i386.deb -> linux, 386
//   - mytool_x86_64-unknown-linux-gnu.tar.xz -> linux, amd64
func AssetNameGuess(name string) (platform string, arch string) {
	base := strings.ToLower(filepath.Base(name))

	// Quick filters for common OS
	switch {
	case strings.Contains(base, "windows") || strings.HasSuffix(base, ".exe"):
		platform = "windows"
	case strings.Contains(base, "darwin") || strings.Contains(base, "mac") || strings.Contains(base, "macos") || strings.Contains(base, "macosx") || strings.Contains(base, "osx"):
		platform = "darwin"
	case strings.Contains(base, "linux"):
		platform = "linux"
	case strings.Contains(base, "freebsd"):
		platform = "freebsd"
	case strings.Contains(base, "openbsd"):
		platform = "openbsd"
	case strings.Contains(base, "netbsd"):
		platform = "netbsd"
	case strings.Contains(base, "android"):
		platform = "android"
	}

	// Architecture patterns
	switch {
	case strings.Contains(base, "arm64") || strings.Contains(base, "aarch64"):
		arch = "arm64"
	case strings.Contains(base, "amd64") || strings.Contains(base, "x86_64"):
		arch = "amd64"
	case strings.Contains(base, "386") || strings.Contains(base, "i386") || strings.Contains(base, "x86-32"):
		arch = "386"
	case strings.Contains(base, "armv7") || strings.Contains(base, "armhf"):
		arch = "arm"
	case strings.Contains(base, "ppc64le"):
		arch = "ppc64le"
	case strings.Contains(base, "ppc64"):
		arch = "ppc64"
	case strings.Contains(base, "s390x"):
		arch = "s390x"
	case strings.Contains(base, "riscv64"):
		arch = "riscv64"
	}

	// Additional heuristic for separators: _, -, .
	if platform == "" || arch == "" {
		tokens := splitTokens(base)
		// Walk tokens to catch patterns like linux amd64
		for _, t := range tokens {
			switch t {
			case "linux", "darwin", "macos", "macosx", "osx", "windows", "win":
				if platform == "" {
					if t == "win" {
						platform = "windows"
					} else if t == "macos" || t == "macosx" || t == "osx" {
						platform = "darwin"
					} else {
						platform = t
					}
				}
			case "amd64", "x86_64":
				if arch == "" {
					arch = "amd64"
				}
			case "arm64", "aarch64":
				if arch == "" {
					arch = "arm64"
				}
			case "386", "i386", "x86-32":
				if arch == "" {
					arch = "386"
				}
			case "armv7", "armhf":
				if arch == "" {
					arch = "arm"
				}
			}
		}
	}

	return platform, arch
}

func splitTokens(s string) []string {
	// Replace separators with spaces and split
	sep := strings.NewReplacer("-", " ", "_", " ", ".", " ")
	return strings.Fields(sep.Replace(s))
}
