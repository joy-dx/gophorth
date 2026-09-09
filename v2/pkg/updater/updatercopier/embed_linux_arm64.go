//go:build linux && arm64
// +build linux,arm64

package updatercopier

import "embed"

//go:embed assets/update-helper-linux-arm64
var embeddedHelpers embed.FS
