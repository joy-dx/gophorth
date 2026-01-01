package cmd

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/joy-dx/gophorth/examples/from-github-release/config"
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/joy-dx/gophorth/pkg/logger"
	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
	"github.com/joy-dx/gophorth/pkg/logger/loggersinks"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/relay"
	"github.com/joy-dx/gophorth/pkg/relay/relayconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserconfig"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			if stateErr := cfg.Process(); stateErr != nil {
				log.Fatal(stateErr)
			}
			cfg.Updater.WithVersion(BuildID)

			// Logger - A simple console relay sink builder
			if viper.GetBool(string(options.Quiet)) {
				cfg.Logger.WithLevel(relaydto.Error)
			}
			if viper.GetBool(string(options.Debug)) {
				cfg.Logger.WithLevel(relaydto.Debug)
				cfg.Logger.WithType(loggersinks.SimpleLoggerRef)
			}
			loggerSvc := logger.ProvideLoggerSvc(&cfg.Logger)
			if err := loggerSvc.Hydrate(); err != nil {
				log.Fatal(fmt.Errorf("problem creating logger: %w", err))
			}
			consoleSink := loggerSvc.GetLogger()

			// Relay - Internal Channel based event bus
			relayCfg := relayconfig.DefaultRelaySvcConfig()
			relaySvc := relay.ProvideRelaySvc(&relayCfg)
			// Register a common screen out sink from the main logger service
			relaySvc.RegisterSink(consoleSink)
			for _, relaySink := range relayCfg.Sinks {
				relaySvc.RegisterSink(relaySink)
			}
			if err := relaySvc.Hydrate(); err != nil {
				log.Fatal(fmt.Errorf("problem creating relay: %w", err))
			}

			// Net - Network operations service with blacklist / whitelist support
			cfg.Net.WithRelay(relaySvc)
			netService := net.ProvideNetSvc(&cfg.Net)
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
	configBuilder.AddBoolParam(options.Debug, false, "Whether to show more output in the console")
	configBuilder.AddBoolParam(options.Quiet, false, "Only show error logs")
	loggerconfig.CobraAndViper(rootCmd)
	netconfig.CobraAndViper(rootCmd)
	relayconfig.CobraAndViper(rootCmd)
	releaserconfig.CobraAndViper(rootCmd)
	updaterdto.CobraAndViper(rootCmd)

}
