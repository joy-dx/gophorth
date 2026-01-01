package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/joy-dx/gophorth/examples/from-github-release/config"
	"github.com/joy-dx/gophorth/examples/utils"
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/relay"
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

			// Create an HTTP client with the custom timeout
			httpClient := &http.Client{
				Timeout: 20 * time.Second,
			}
			githubAgent := github.NewClient(httpClient)

			githubClientCfg := updaterclients.DefaultFromGithubConfig()
			githubClientCfg.WithOwner("joy-dx").
				WithRepo("app-update-example").
				WithClient(githubAgent)
			githubClient := updaterclients.NewFromGithub(&githubClientCfg)

			// Update Client
			logPath, logPathErr := filepath.Abs("./update.log")
			if logPathErr != nil {
				log.Fatal(logPathErr)
			}
			cfgSvc.Updater.WithRelay(relaySvc).
				WithNetSvc(netSvc).
				WithCheckClient(githubClient).
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
