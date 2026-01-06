package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"wails-app/config"

	"github.com/joy-dx/gonetic"
	netCfg "github.com/joy-dx/gonetic/config"
	"github.com/joy-dx/gophorth/pkg/releaser"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay"
	relayCfg "github.com/joy-dx/relay/config"
	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/sinks"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prepare-assets <version>")
		os.Exit(1)
	}

	versionWanted := os.Args[1]

	ctx := context.Background()
	cfg := config.ProvideConfigSvc()
	cfg.Updater = updaterdto.DefaultUpdaterSvcConfig()
	cfg.Net = netCfg.DefaultNetSvcConfig()
	cfg.Relay = relayCfg.DefaultRelaySvcConfig()
	cfg.Releaser = releaserdto.DefaultReleaserConfig()
	if stateErr := cfg.Process(); stateErr != nil {
		log.Fatal(stateErr)
	}

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
	netSvc := gonetic.ProvideNetSvc(&cfg.Net)
	if err := netSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating net service: %w", err))
	}

	cfg.Releaser.WithRelay(relaySvc).
		WithNetSvc(netSvc).
		WithVersion(versionWanted).
		WithOutputPath("./assets").
		WithPrivateKeyPath("./embedded/private-pgp.key").
		WithTargetPath("./assets").
		WithFilePattern("wails-app-{platform}-{arch}{variant}").
		WithDownloadPrefix("http://localhost:8080/").
		WithAllowAnyExtension(true)

	releaserSvc := releaser.ProvideReleaserSvc(&cfg.Releaser)
	if err := releaserSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating releaser: %w", err))
	}

	if _, err := releaserSvc.GenerateReleaseSummary(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem generating release summary: %w", err))
	}
}
