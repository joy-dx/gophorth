package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"wails-app/config"
	"wails-app/guisinks"

	"github.com/Masterminds/semver/v3"
	"github.com/joy-dx/gophorth/examples/utils"
	"github.com/joy-dx/gophorth/pkg/archive"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater"
	"github.com/joy-dx/gophorth/pkg/updater/updaterclients"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay"
	relayCfg "github.com/joy-dx/relay/config"
	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
	"github.com/joy-dx/relay/sinks"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	BuildID = "0.0.1"
	Variant = ""
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	workingDir, err := getWorkingDir()
	if err != nil {
		log.Fatal(err)
	}

	// Serve the assets path update checking

	cfg := config.ProvideConfigSvc()
	cfg.Updater = updaterdto.DefaultUpdaterSvcConfig()
	cfg.Net = netconfig.DefaultNetSvcConfig()
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
	netSvc := net.ProvideNetSvc(&cfg.Net)
	if err := netSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating net service: %w", err))
	}

	// CHECK CLIENT - net type to download meta information from an endpoint
	netClientCfg := updaterclients.DefaultFromNetConfig()
	netClientCfg.WithUserFetchFunction(func(ctx context.Context, cfg updaterclients.NetAgentCfg) (releaserdto.ReleaseAsset, error) {

		var releaseSummary releaserdto.ReleaseSummary
		var releaseAsset releaserdto.ReleaseAsset
		url := "http://localhost:8080/version.json"
		if response, err := cfg.NetSvc.Get(ctx, url, true); err != nil {
			return releaseAsset, err
		} else {
			if unmarshalErr := json.Unmarshal(response.Body, &releaseSummary); unmarshalErr != nil {
				return releaseAsset, unmarshalErr
			}
		}

		remoteVersionSemVer, remoteVersionErr := semver.NewVersion(releaseSummary.Version)
		if remoteVersionErr != nil {
			cfg.Relay.Warn(updater.RlyUpdaterLog{Msg: fmt.Sprintf("Couldn't parse remote version: %s. %s", releaseSummary.Version, remoteVersionErr.Error())})
			return releaseAsset, fmt.Errorf("couldn't parse remote version: %w", remoteVersionErr)
		}
		cfg.Relay.Debug(updater.RlyUpdaterLog{Msg: fmt.Sprintf("Remote version found: %s", remoteVersionSemVer.String())})
		for _, asset := range releaseSummary.Assets {
			if asset.Platform == cfg.UpdaterCfg.Platform && asset.Arch == cfg.UpdaterCfg.Architecture {
				cfg.Relay.Debug(updater.RlyUpdaterLog{Msg: fmt.Sprintf("found asset with matching platform: %s %s", asset.Platform, asset.Arch)})
				if cfg.UpdaterCfg.Variant == asset.Variant {
					cfg.Relay.Debug(updater.RlyUpdaterLog{Msg: fmt.Sprintf("found wanted variant %s", asset.Variant)})
					return asset, nil
				}
			}

		}
		return releaseAsset, errors.New("couldn't find remote version")
	})
	netClient := updaterclients.NewFromNet(&netClientCfg)

	prepareFunc := func(ctx context.Context, cfg *updaterdto.UpdaterAgentCfg) error {
		// On mac we distribute the app as an archive
		// unpack and update the artefact path
		if runtime.GOOS == "darwin" {
			extractOptions := archive.DefaultExtractOptions()
			err := archive.Extract(ctx, cfg.VersionUpdate.ArtefactName, cfg.UpdaterCfg.TemporaryPath, extractOptions)
			if err != nil {
				return err
			}

			// Update the artefact path
			cfg.VersionUpdate.WithArtefactName(cfg.UpdaterCfg.TemporaryPath + "/wails-app.app")
			return nil
		}

		log.Println("running custom update function")
		return nil
	}

	// Update Client
	logPath, logPathErr := filepath.Abs(workingDir + "/update.log")
	relaySvc.Debug(events.RlyLog{Msg: fmt.Sprintf("using %s as update log path", logPath)})
	if logPathErr != nil {
		log.Fatal(logPathErr)
	}
	cfg.Updater.WithRelay(relaySvc).
		WithNetSvc(netSvc).
		WithCheckClient(netClient).
		WithTemporaryPath("/tmp/update-test").
		WithVersion(BuildID).
		WithPublicKeyPath(workingDir + "/embedded/public-pgp.key").
		WithUpdateLogPath(logPath).
		WithPrepareFunc(prepareFunc).
		WithVariant(Variant)

	updaterSvc := updater.ProvideUpdaterSvc(&cfg.Updater)
	if err := updaterSvc.Hydrate(ctx); err != nil {
		log.Fatal(fmt.Errorf("problem creating updater service: %w", err))
	}
	updaterInterface := &UpdaterInterface{
		updaterSvc: updaterSvc,
		relay:      relaySvc,
	}

	// Create application with options
	wailsErr := wails.Run(&options.App{
		Title:  "wails-app",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			updaterInterface.SetContext(ctx)

			wailsSink := guisinks.NewWailsSink(ctx, &dto.RelaySinkConfig{
				Level: dto.Debug,
				Ref:   "wails",
			})
			relaySvc.RegisterSink(wailsSink)
			relaySvc.Info(events.RlyLog{Msg: fmt.Sprintf("hosting assets from %s/assets", workingDir)})

			// Host the updates path so we can reach out to the web
			go func() {
				if err := utils.ServeDir(ctx, "localhost:8080", workingDir+"/assets"); err != nil {
					log.Fatal(err)
				}
			}()

		},
		Bind: []interface{}{
			updaterInterface,
		},
		EnumBind: []interface{}{
			Channels,
			Relays,
		},
	})

	if wailsErr != nil {
		println("Error:", wailsErr.Error())
	}
}

// getWorkingDir Included in the demo app as a little hack to speedily get paths working
func getWorkingDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	exePath := filepath.Dir(exe)

	// If running on mac (within a .app) skip up to directory levels
	if runtime.GOOS == "darwin" {
		exePath = filepath.Dir(exePath)
		exePath = filepath.Dir(exePath)
		exePath = filepath.Dir(exePath)
	}

	return exePath, nil

}
