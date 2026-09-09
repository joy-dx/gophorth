//go:build linux && amd64
// +build linux,amd64

package updatercopier

import "embed"

//go:embed assets/update-helper-linux-amd64
var embeddedHelpers embed.FS
