package cmd

import (
	"fmt"

	"github.com/joy-dx/gophorth/examples/from-json-url/config"
	"github.com/joy-dx/gophorth/pkg/updater"
	"github.com/joy-dx/relay"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			cfgSvc := config.ProvideConfigSvc()
			relaySvc := relay.ProvideRelaySvc(nil)
			relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("Current version is: %s", cfgSvc.Updater.Version)})
		},
	}
)

func init() { //nolint:gochecknoinits // using cobra
	rootCmd.AddCommand(versionCmd)
}
