package cmd

import (
	"fmt"
	"log"

	"github.com/joy-dx/gophorth/examples/from-json-url/config"
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/relay"
	"github.com/joy-dx/gophorth/pkg/releaser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	prepareAssetsCmd = &cobra.Command{
		Use: "prepare-assets",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cfgSvc := config.ProvideConfigSvc()
			relaySvc := relay.ProvideRelaySvc(nil)
			netSvc := net.ProvideNetSvc(nil)

			versionWanted := viper.GetString(string(options.ReleaserVersion))

			cfgSvc.Releaser.WithRelay(relaySvc).
				WithNetSvc(netSvc).
				WithVersion(versionWanted).
				WithOutputPath("./cmd/assets").
				WithPrivateKeyPath("./cmd/embedded/private-pgp.key").
				WithTargetPath("./cmd/assets").
				WithFilePattern("app-example-{platform}-{arch}").
				WithDownloadPrefix("http://localhost:8080/")
			releaserSvc := releaser.ProvideReleaserSvc(&cfgSvc.Releaser)
			if err := releaserSvc.Hydrate(ctx); err != nil {
				log.Fatal(fmt.Errorf("problem creating releaser: %w", err))
			}

			if _, err := releaserSvc.GenerateReleaseSummary(ctx); err != nil {
				log.Fatal(fmt.Errorf("problem generating release summary: %w", err))
			}
		},
	}
)

func init() { //nolint:gochecknoinits // using cobra
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(prepareAssetsCmd)
	configBuilder.AddStringParam(options.ReleaserVersion, "1.0.0", "What version is being released if not found in file information")

	rootCmd.AddCommand(prepareAssetsCmd)
}
