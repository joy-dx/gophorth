//go:build darwin && amd64
// +build darwin,amd64

package updatercopier

import "embed"

//go:embed assets/update-helper-darwin-amd64
var embeddedHelpers embed.FS
