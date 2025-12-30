package updaterdto

import (
	"runtime"
	"time"

	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/spf13/cobra"
)

const ConfigPrefix = "updater"

func CobraAndViper(cmd *cobra.Command) {
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(cmd)
	configBuilder.SetConfigPrefix([]string{ConfigPrefix})
	configBuilder.AddBoolParam(options.UpdaterAllowDowngrade, false, "allows downgrading to older versions")
	configBuilder.AddBoolParam(options.UpdaterAllowPrerelease, false, "allows updating to pre-release versions")
	configBuilder.AddStringParam(options.UpdaterArchitecture, runtime.GOARCH, "If no conforming to GOOS standards, string representing architecture part")
	configBuilder.AddDurationParam(options.UpdaterCheckInterval, 48*time.Hour, "How often to check for latest JoyDX version")
	configBuilder.AddStringParam(options.UpdaterCurrentVersion, "0.0.1", "Semantic version representing current runtime version")
	configBuilder.AddStringParam(options.UpdaterLogPath, ".", "Local file system path used during update as log path")
	configBuilder.AddStringParam(options.UpdaterPlatform, runtime.GOOS, "If no conforming to GOOS standards, string representing platform part")
	configBuilder.AddStringParam(options.UpdaterPublicKey, "", "Contains ASCII encode EDCSA or PGP public key")
	configBuilder.AddStringParam(options.UpdaterPublicKeyPath, "", "Path to EDCSA or PGP public key")
	configBuilder.AddStringParam(options.UpdaterTemporaryPath, "./tmp", "Where to store download and update artefacts")
	configBuilder.AddStringParam(options.UpdaterVariant, "", "Represents a download variant that the current device wants")
}
