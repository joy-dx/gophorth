package cmd

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/joy-dx/gonetic"
	netCfg "github.com/joy-dx/gonetic/config"
	"github.com/joy-dx/gophorth/examples/from-github-release/config"
	"github.com/joy-dx/gophorth/examples/from-github-release/config/cliflags"
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserconfig"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay"
	relayCfg "github.com/joy-dx/relay/config"
	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/sinks"
	"github.com/spf13/cobra"
)

//go:embed embedded/*
var assets embed.FS

var (
	BuildID                              = "0.0.1"
	cfgFile                              string
	cancelContext, cancelContextFunction = context.WithCancel(context.Background())
	rootCmd                              = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(cancelContext)
			cfg := config.ProvideConfigSvc()
			cfg.Updater = updaterdto.DefaultUpdaterSvcConfig()
			cfg.Net = netCfg.DefaultNetSvcConfig()
			cfg.Relay = relayCfg.DefaultRelaySvcConfig()
			cfg.Releaser = releaserdto.DefaultReleaserConfig()
			if stateErr := cfg.Process(); stateErr != nil {
				log.Fatal(stateErr)
			}
			cfg.Updater.WithVersion(BuildID)

			consoleCfg := sinks.DefaultSimpleLoggerConfig()
			consoleCfg.WithLevel(dto.Debug)
			consoleSink := sinks.NewSimpleLogger(&consoleCfg)

			// Relay - Internal Channel based event bus
			relaySvc := relay.ProvideRelaySvc(&cfg.Relay)
			// Register a common screen out sink from the main logger service
			relaySvc.RegisterSink(consoleSink)
			for _, relaySink := range cfg.Relay.Sinks {
				relaySvc.RegisterSink(relaySink)
			}
			if err := relaySvc.Hydrate(); err != nil {
				log.Fatal(fmt.Errorf("problem creating relay: %w", err))
			}

			// Net - Network operations service with blacklist / whitelist support
			cfg.Net.WithRelay(relaySvc)
			netService := gonetic.ProvideNetSvc(&cfg.Net)
			if err := netService.Hydrate(cancelContext); err != nil {
				log.Fatal(fmt.Errorf("problem creating net service: %w", err))
			}
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, string(options.ConfigPath), "$HOME/.joydx.yaml",
		"config file path")

	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(rootCmd)
	cliflags.NetCobraAndViper(rootCmd)
	releaserconfig.CobraAndViper(rootCmd)
	updaterdto.CobraAndViper(rootCmd)

}
