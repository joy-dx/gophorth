# GoPhorth

An unopinionated Go app update mechanism loaded with all the tooling you need to make self-updating Go apps a breeze. 

* Releaser helps prepare deployment artefacts with checksum and signature generation
* Updater provides a light to implement but fully customisable and reliable app self-update mechanism
* File replacement delegated to external process for reliability
* Support tooling baked in to turn activities such as accessing Github, fetching net assets or decompressing archives in to one liners

The reason for building this is that I found existing solutions were highly opinionated with very limited flexibility. Using Wails as an example, on Linux you will likely be distributing just the binary but, for MacOS, you will be distributing a `.app` in an archive. The running process is a file within the `.app` directory so needs some special handling.

## Quickstart

Depending on your specific needs there are some ready-made demonstration apps in the `examples` folder you are recommended to review for exact implementation details.

```
cd examples/from-json-url
# Generate relevant artefacts for demonstration. e.g. version 1 in project root and version 2 in 
# assets path with checksums, signatures, and release meta data
make generate

# Run version 1 of the app and check for update
./app-example update

# After the update, try running update again
./app-example update
```

## Getting Started

### Preparing your artefacts

GoPhorth comes with a Releaser service to help prepare meta information about your published artefacts. You can see an [example file here](examples/version-information.json)

```go
// Internal messaging system to extend log capabilities
relaySvc := relay.ProvideRelaySvc(nil)
// Network services facade
netSvc := gonetic.ProvideNetSvc(nil)

releaserCfg := releaserdto.DefaultReleaserConfig()
releaserCfg.WithRelay(relaySvc).
    WithNetSvc(netSvc).
    // Manually specify version if not included in the file name
	WithVersion("2.0.0").
	// Artefacts will be placed here
    WithOutputPath("./assets").
	// If generating (PGP or ECDSA) signatures, provide path to the private key 
    WithPrivateKeyPath("~/secrets/private-pgp.key").
	// Path to directory where published artefacts are located
    WithTargetPath("./assets").
	// Specify file name pattern for files to process and way to extract information
    WithFilePattern("app-example-{platform}-{arch}").
	// If you already know where URL artefacts will be hosted
    WithDownloadPrefix("http://localhost:8080/")

releaserSvc := releaser.ProvideReleaserSvc(&cfgSvc.Releaser)
if err := releaserSvc.Hydrate(ctx); err != nil {
    log.Fatal(fmt.Errorf("problem creating releaser: %w", err))
}

if _, err := releaserSvc.GenerateReleaseSummary(ctx); err != nil {
    log.Fatal(fmt.Errorf("problem generating release summary: %w", err))
}
```

For running the above, in your output path, you will get:

* `version.json` provides a release and asset summary for programmatic consumption including checksums and signatures
* `checksums.txt` provides SHA256 hashes of the processed files for verification
* `FILE_NAME.asc` provides signatures for each processed file for verification

### Checking and initiating updates

```go
// Update check workflow using a web endpoint to fetch release info
netClientCfg := updaterclients.DefaultFromNetConfig()
netClientCfg.WithUserFetchFunction(func(ctx context.Context, cfg updaterclients.NetAgentCfg) (releaserdto.ReleaseAsset, error) {
    // ReleaseSummary and ReleaseAsset are the internal structs used
	// In this case, they are also what gets returned from the demo endpoint
	// Switch these out for your endpoint model
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

	// Convert the found version in to a semantic version struct for checking
    remoteVersionSemVer, remoteVersionErr := semver.NewVersion(releaseSummary.Version)
    if remoteVersionErr != nil {
        return releaseAsset, fmt.Errorf("couldn't parse remote version: %w", remoteVersionErr)
    }
	
	// Refine selection from response to match current device
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

// Set a predictable destination for the update log. it is used post-update to
// provide a success report
logPath, logPathErr := filepath.Abs("./update.log")
if logPathErr != nil {
    log.Fatal(logPathErr)
}

// Initiate the updater service
updaterCfg := updaterdto.DefaultUpdaterSvcConfig()
updaterCfg.WithRelay(relaySvc).
    WithNetSvc(netSvc).
    WithCheckClient(netClient).
    WithTemporaryPath("/tmp/update-test").
    // Provide the current app version for comparison
    WithVersion("1.0.0"). 
	// If releases are signed, provide public key part for verification
    WithPublicKeyPath("./cmd/embedded/public-pgp.key").
    WithUpdateLogPath(logPath)

updaterSvc := updater.ProvideUpdaterSvc(&updaterCfg)
if err := updaterSvc.Hydrate(ctx); err != nil {
    log.Fatal(fmt.Errorf("problem creating updater service: %w", err))
}

// Get the latest version from remote
latestVersion, err := updaterSvc.CheckUpdate(ctx)
if err != nil {
    log.Fatal(fmt.Errorf("problem checking for latest version: %w", err))
}
switch updaterSvc.Status() {
// Update handling
case updaterdto.UPDATE_AVAILABLE:
    
	if downloadErr := updaterSvc.DownloadUpdate(ctx, nil); downloadErr != nil {
        log.Fatal(fmt.Errorf("problem downloading latest version: %w", downloadErr))
    }
    if updateErr := updaterSvc.PerformUpdate(ctx); updateErr != nil {
        log.Fatal(fmt.Errorf("problem performing update: %w", updateErr))
    }
case updaterdto.UP_TO_DATE:
    relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("already up to date: %s", latestVersion.Version)})
default:
    relaySvc.Info(updater.RlyUpdaterLog{Msg: fmt.Sprintf("unhandled download state: %s", updaterSvc.Status())})
}
```

## The Update Workflow

for reference, the following occurs once `PerformUpdate` is called

* A helper places an update-helper to the temporary path
* The update helper is started as a separate process using update target (current program), update artefact path, and a log file path as arguments. The helper:
  * Creates a backup
  * Attempts to remove the update target
  * Replace the update target with the new artefact
  * Attempt to launch new artefact
    * If update fails, rollback
  * Cleanup the backup
* Upon updater service hydration, the app reads the update log
* Set update status to complete

For any followup operations or reading the update log, you can extend the following

```go
if updaterSvc.Status() == updaterdto.COMPLETE {
    relaySvc.Info(updater.RlyUpdaterLog{Msg: updaterSvc.UpdateLog()})
    // PostInstallCleanup removes the log file
    if err := updaterSvc.PostInstallCleanup(); err != nil {
        relaySvc.Warn(updater.RlyUpdaterLog{Msg: err.Error()})
    }
}
```