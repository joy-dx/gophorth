package releaserconfig

import (
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/spf13/cobra"
)

const ConfigPrefix = "releaser"

func CobraAndViper(cmd *cobra.Command) {
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(cmd)
	configBuilder.SetConfigPrefix([]string{ConfigPrefix})
	configBuilder.AddStringParam(options.ReleaserOutputPath, "./out", "FS Path where generated artefacts will be saved")
	configBuilder.AddStringParam(options.ReleaserTargetPath, "./cmd/assets", "FS path to published artefacts")
	configBuilder.AddStringParam(options.ReleaserFilePattern, "app-example-{platform}-{arch}", "name the published app to be processed starts with")
	configBuilder.AddStringParam(options.ReleaserPrivateKey, "", "Contains ASCII encode EDCSA or PGP public key")
	configBuilder.AddStringParam(options.ReleaserPrivateKeyPath, "./cmd/assets/private-pgp.key", "Path to EDCSA or PGP public key")
	configBuilder.AddBoolParam(options.ReleaserAllowAnyExtension, false, "Allows a file extension after the pattern (\".zip\", \".tar.gz\", etc.)")
	configBuilder.AddBoolParam(options.ReleaserStrict, false, "If true, non-matching files cause an error. If false, they are skipped.")
	configBuilder.AddBoolParam(options.ReleaserRequireVersion, false, "If true, {version} is treated as required when used in the pattern.")
	configBuilder.AddBoolParam(options.ReleaserGenerateChecksums, true, "Whether to create a separate checksums.txt artefact")
	configBuilder.AddBoolParam(options.ReleaserGenerateSignatures, true, "If available, create signatures of the artefacts and store in ASCII armored format")
	configBuilder.AddStringParam(options.ReleaserSummaryOutputType, "json-indented", "Format to output the summary file in")
	configBuilder.AddStringParam(options.ReleaserVersion, "0.0.1", "Manually specify version to use with release")
}
