//go:build windows && amd64
// +build windows,amd64

package updatercopier

import "embed"

//go:embed assets/update-helper-windows-amd64.exe
var embeddedHelpers embed.FS
