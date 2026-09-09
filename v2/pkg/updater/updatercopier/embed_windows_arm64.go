//go:build windows && arm64
// +build windows,arm64

package updatercopier

import "embed"

//go:embed assets/update-helper-windows-arm64.exe
var embeddedHelpers embed.FS
