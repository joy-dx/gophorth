package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/file"
	"github.com/joy-dx/gophorth/pkg/hydrate"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

func (s *UpdaterSvc) UpdateLink() *releaserdto.ReleaseAsset {
	return s.contextUpdate
}

func (s UpdaterSvc) State() *updaterdto.UpdaterState {
	return &updaterdto.UpdaterState{
		LastTimeCheckedUpdate: s.cfg.LastUpdateCheck,
		UpdateLink:            s.contextUpdate,
		Changelog:             s.changelog,
		ReleasedAt:            s.releasedAt,
		CheckInterval:         s.cfg.CheckInterval,
		Log:                   s.updateLog,
		LogPath:               s.cfg.LogPath,
		Version:               s.version.String(),
	}
}

func (s UpdaterSvc) Status() updaterdto.UpdateStatus {
	return s.status
}

func (s UpdaterSvc) UpdateLog() string {
	return s.updateLog
}

func (s *UpdaterSvc) Hydrate(ctx context.Context) error {
	if err := hydrate.NilCheck("net", map[string]interface{}{
		"config": s.cfg,
		"relay":  s.relay,
	}); err != nil {
		return err
	}
	s.relay.Debug(RlyUpdaterLog{Msg: "start: hydrate state"})

	// Is the app finishing an upgrade?
	if s.cfg.LogPath != "" {
		logContents, readErr := file.ToBytes(s.cfg.LogPath)
		if readErr == nil {
			s.updateLog = string(logContents)
			s.status = updaterdto.COMPLETE
		}
	}

	if s.cfg.PublicKeyPath != "" {
		s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("public key provided from file: %s", s.cfg.PublicKeyPath)})
		publicKey, err := file.ToBytes(s.cfg.PublicKeyPath)
		if err != nil {
			s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("failed to read public key file %s. %w", s.cfg.PublicKeyPath, err.Error())})
		}
		s.cfg.WithPublicKey(string(publicKey))
	}

	if s.cfg.PublicKey != "" {
		s.relay.Debug(RlyUpdaterLog{Msg: "public key provided for checking releases against"})
		if keyInfo, err := cryptography.DetectSignatureInformation([]byte(s.cfg.PublicKey)); err != nil {
			s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("could not detect key information: %s", err.Error())})
		} else {
			switch keyInfo.Format {
			case "PGP":
				pgpEntity, keyringErr := cryptography.LoadKeyRingAuto([]byte(s.cfg.PublicKey))
				if keyringErr != nil {
					s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("could not load public key information: %s", keyringErr.Error())})
				}
				s.pgpEntity = pgpEntity
			case "X509":
				ecdsaKey, keyringErr := cryptography.ParseECDSAPublicKeyFromPEM(s.cfg.PublicKey)
				if keyringErr != nil {
					s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("could not load public key information: %s", keyringErr.Error())})
				}
				s.ecdsaKey = ecdsaKey
			default:
				s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("unsupported key format: %s", keyInfo.Format)})
			}
		}
	}

	if s.cfg.Version != "" {
		parsedVersion, err := semver.NewVersion(s.cfg.Version)
		if err != nil {
			return fmt.Errorf("could not parse release version: %w", err)
		}
		s.version = parsedVersion
	}

	needUpdateCheck := true
	if s.cfg.LastUpdateCheck != nil {
		now := time.Now()
		threshold := now.Add(-s.cfg.CheckInterval)
		s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("last time checked update: %s. Should check at: %s", s.cfg.LastUpdateCheck.String(), threshold.String())})
		if s.cfg.LastUpdateCheck.After(threshold) {
			needUpdateCheck = false
		}
	}

	if needUpdateCheck {
		s.relay.Debug(RlyUpdaterLog{Msg: "update has crossed check interval, checking in BG"})
		go func() {
			if _, checkErr := s.CheckLatest(ctx); checkErr != nil {
				s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("problem checking update in BG: %s", checkErr.Error())})

			}
		}()
	} else {
		s.relay.Debug(RlyUpdaterLog{Msg: "no need to check for updates"})
	}

	selfPath, err := os.Executable()
	if err != nil {
		return err
	}
	updateTarget := selfPath
	if s.cfg.Platform == "darwin" {
		appBundleRoot := findAppBundleRoot(selfPath)
		if appBundleRoot != "" {
			s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("found app bundle at %s", appBundleRoot)})
			updateTarget = appBundleRoot
		}
	}
	s.updateTarget = updateTarget

	s.relay.Debug(RlyUpdaterLog{Msg: "end: hydrate state"})
	return nil
}

func (s *UpdaterSvc) PostInstallCleanup() error {
	s.relay.Debug(RlyUpdaterLog{Msg: "Post install cleanup"})
	removeErr := os.Remove(s.cfg.LogPath)
	if removeErr != nil {
		s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("failed to remove log file %s", removeErr.Error())})
	}
	s.cfg.WithUpdateLogPath("")
	return nil
}

// findAppBundleRoot walks up the path to locate the parent .app directory.
func findAppBundleRoot(p string) string {
	parts := strings.Split(p, string(os.PathSeparator))
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasSuffix(parts[i], ".app") {
			return "/" + filepath.Join(parts[:i+1]...)
		}
	}
	return ""
}
