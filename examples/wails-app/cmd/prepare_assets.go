package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"wails-app/config"

	gophoptions "github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/joy-dx/gophorth/pkg/logger"
	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
	"github.com/joy-dx/gophorth/pkg/logger/loggersinks"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/relay"
	"github.com/joy-dx/gophorth/pkg/relay/relayconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/joy-dx/gophorth/pkg/releaser"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/spf13/viper"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prepare-assets <version>")
		os.Exit(1)
	}

	versionWanted := os.Args[1]

	ctx := context.Background()
	cfgSvc := config.ProvideConfigSvc()
	cfgSvc.Logger = loggerconfig.DefaultLoggerConfig()
	cfgSvc.Updater = updaterdto.DefaultUpdaterSvcConfig()
	cfgSvc.Net = netconfig.DefaultNetSvcConfig()
	cfgSvc.Relay = relayconfig.DefaultRelaySvcConfig()
	cfgSvc.Releaser = releaserdto.DefaultReleaserConfig()
	if stateErr := cfgSvc.Process(); stateErr != nil {
		log.Fatal(stateErr)
	}
	// Logger - A simple console relay sink builder
	if viper.GetBool(string(gophoptions.Quiet)) {
		cfgSvc.Logger.WithLevel(relaydto.Error)
	}
	if viper.GetBool(string(gophoptions.Debug)) {
		cfgSvc.Logger.WithLevel(relaydto.Debug)
		cfgSvc.Logger.WithType(loggersinks.SimpleLoggerRef)
	}
	loggerSvc := logger.ProvideLoggerSvc(&cfgSvc.Logger)
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
	cfgSvc.Net.WithRelay(relaySvc)
	netSvc := net.ProvideNetSvc(&cfgSvc.Net)
	if err := netSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating net service: %w", err))
	}

	cfgSvc.Releaser.WithRelay(relaySvc).
		WithNetSvc(netSvc).
		WithVersion(versionWanted).
		WithOutputPath("./assets").
		WithPrivateKeyPath("./embedded/private-pgp.key").
		WithTargetPath("./assets").
		WithFilePattern("wails-app-{platform}-{arch}{variant}").
		WithDownloadPrefix("http://localhost:8080/").
		WithAllowAnyExtension(true)

	releaserSvc := releaser.ProvideReleaserSvc(&cfgSvc.Releaser)
	if err := releaserSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating releaser: %w", err))
	}

	if _, err := releaserSvc.GenerateReleaseSummary(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem generating release summary: %w", err))
	}
}
