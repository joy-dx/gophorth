//go:build darwin && arm64
// +build darwin,arm64

package updatercopier

import "embed"

//go:embed assets/update-helper-darwin-arm64
var embeddedHelpers embed.FS
