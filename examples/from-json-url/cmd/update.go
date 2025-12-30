package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joy-dx/gophorth/examples/from-json-url/config"
	"github.com/joy-dx/gophorth/examples/utils"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/relay"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater"
	"github.com/joy-dx/gophorth/pkg/updater/updaterclients"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use: "update",
		Run: func(cmd *cobra.Command, args []string) {

			cfgSvc := config.ProvideConfigSvc()
			relaySvc := relay.ProvideRelaySvc(nil)
			netSvc := net.ProvideNetSvc(nil)

			relaySvc.Info(updater.RlyUpdaterLog{Msg: "Starting updater"})
			// Root context cancelled on SIGINT/SIGTERM
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			// Serve the assets path update checking
			go func() {
				if err := utils.ServeDir(ctx, "localhost:8080", "./cmd/assets"); err != nil {
					log.Fatal(err)
				}
			}()

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

			// Update Client
			logPath, logPathErr := filepath.Abs("./update.log")
			if logPathErr != nil {
				log.Fatal(logPathErr)
			}
			cfgSvc.Updater.WithRelay(relaySvc).
				WithNetSvc(netSvc).
				WithCheckClient(netClient).
				WithTemporaryPath("/tmp/update-test").
				WithVersion(BuildID).
				WithPublicKeyPath("./cmd/embedded/public-pgp.key").
				WithUpdateLogPath(logPath)

			updaterSvc := updater.ProvideUpdaterSvc(&cfgSvc.Updater)
			if err := updaterSvc.Hydrate(ctx); err != nil {
				log.Fatal(fmt.Errorf("problem creating updater service: %w", err))
			}

			if updaterSvc.Status() == updaterdto.COMPLETE {
				relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("app update complete. now on: %s", BuildID)})
				updateLog := updaterSvc.UpdateLog()
				if updateLog != "" {
					relaySvc.Info(updater.RlyUpdaterLog{Msg: updateLog})
				}
				if err := updaterSvc.PostInstallCleanup(); err != nil {
					relaySvc.Warn(updater.RlyUpdaterLog{Msg: err.Error()})
				}
				os.Exit(0)
			}

			relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("current app version is: %s", BuildID)})
			relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("looking for %s %s", cfgSvc.Updater.Platform, cfgSvc.Updater.Architecture)})

			latestVersion, err := updaterSvc.CheckLatest(ctx)
			if err != nil {
				log.Fatal(fmt.Errorf("problem checking for latest version: %w", err))
			}
			switch updaterSvc.Status() {
			case updaterdto.UPDATE_AVAILABLE:
				relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("new version available: %s", latestVersion.Version)})
				relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("Source: %s", latestVersion.DownloadURL)})
				if latestVersion.Checksum != "" {
					relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("checksum available: %s", latestVersion.Checksum)})
				}
				if latestVersion.Signature != "" {
					relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("%s signature is available", latestVersion.SignatureType)})
				}
				time.Sleep(1 * time.Second)
				relaySvc.Info(updater.RlyUpdaterLog{Msg: "Press return to start..."})
				var dummy string
				fmt.Scanln(&dummy)

				// Uses the discovered version link during check to fetch and automatically verify the download
				// via SHA checksums if available and also signature (in this cases PGP)
				if downloadErr := updaterSvc.DownloadUpdate(ctx, nil); downloadErr != nil {
					log.Fatal(fmt.Errorf("problem downloading latest version: %w", downloadErr))
				}

				relaySvc.Info(updater.RlyUpdaterLog{Msg: "Applying update"})
				if updateErr := updaterSvc.PerformUpdate(ctx); updateErr != nil {
					log.Fatal(fmt.Errorf("problem performing update: %w", updateErr))
				}

			case updaterdto.UP_TO_DATE:
				relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("already up to date: %s", latestVersion.Version)})
			default:
				relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("unhandled download state: %s", updaterSvc.Status())})
			}
		},
	}
)

func init() { //nolint:gochecknoinits // using cobra
	rootCmd.AddCommand(updateCmd)
}
