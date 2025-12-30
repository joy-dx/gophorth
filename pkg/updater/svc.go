package updater

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updatercopier"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

type UpdaterSvc struct {
	relay  relaydto.RelayInterface
	netSvc netdto.NetInterface
	cfg    *updaterdto.UpdaterConfig
	status updaterdto.UpdateStatus
	// State information to be populated about possible update
	updateLog     string
	updateTarget  string
	changelog     string
	pgpEntity     openpgp.EntityList
	ecdsaKey      *ecdsa.PublicKey
	releasedAt    *time.Time
	releaseURL    string
	version       *semver.Version
	contextUpdate *releaserdto.ReleaseAsset
}

func (s *UpdaterSvc) CheckLatest(ctx context.Context) (releaserdto.ReleaseAsset, error) {

	if s.cfg.CheckClient == nil {
		return releaserdto.ReleaseAsset{}, errors.New("no check client configured")
	}
	remoteUpdate, err := s.cfg.CheckClient.CheckUpdate(ctx, s.cfg)
	if err != nil {
		return releaserdto.ReleaseAsset{}, fmt.Errorf("check client: %w", err)
	}

	remoteSemVer, err := semver.NewVersion(remoteUpdate.Version)
	if err != nil {
		log.Fatal(fmt.Errorf("problem parsing latest version: %w", err))
	}
	s.relay.Debug(RlyUpdaterLog{Msg: fmt.Sprintf("current version / remote version: %s / %s", s.version, remoteSemVer.String())})
	if remoteSemVer.GreaterThan(s.version) {
		s.status = updaterdto.UPDATE_AVAILABLE
	} else {
		s.status = updaterdto.UP_TO_DATE
	}

	s.contextUpdate = &remoteUpdate
	return remoteUpdate, nil
}

func (s *UpdaterSvc) DownloadUpdate(ctx context.Context, link *releaserdto.ReleaseAsset) error {
	if link != nil {
		s.contextUpdate = link
	}

	var downloadDestination string
	if s.cfg.DownloadFunc != nil {
		agentConfig := updaterdto.UpdaterAgentCfg{
			NetSvc:        s.netSvc,
			UpdaterCfg:    updaterdto.UpdaterConfig{},
			VersionUpdate: s.contextUpdate,
		}
		downloadPath, downloadErr := s.cfg.DownloadFunc(ctx, &agentConfig)
		if downloadErr != nil {
			return downloadErr
		}
		downloadDestination = downloadPath

	} else {
		if s.contextUpdate.DownloadURL == "" {
			return errors.New("no download url configured")
		}
		downloadCfg := netdto.DefaultDownloadFileConfig()
		downloadCfg.WithURL(s.contextUpdate.DownloadURL).
			WithChecksum(s.contextUpdate.Checksum).
			WithDestinationFolder(s.cfg.TemporaryPath)
		downloadPath, downloadErr := s.netSvc.DownloadFile(ctx, &downloadCfg)
		if downloadErr != nil {
			return downloadErr
		}
		downloadDestination = downloadPath
	}
	s.contextUpdate.WithArtefactName(downloadDestination)

	// Ensure the download is executable
	if modErr := os.Chmod(downloadDestination, 0770); modErr != nil {
		return modErr
	}

	if s.contextUpdate.Signature != "" {
		keyInfo, err := cryptography.DetectSignatureInformation([]byte(s.contextUpdate.Signature))
		if err != nil {
			return fmt.Errorf("could not detect key information from link signature: %w", err)
		}
		switch keyInfo.Format {
		case "PGP":
			if s.pgpEntity == nil {
				s.relay.Debug(RlyUpdaterLog{Msg: "pgp signature provided but no local handler"})
			} else {
				signatureAsBuffer := bytes.NewBufferString(s.contextUpdate.Signature)
				if verifyErr := cryptography.PGPVerifyFile(s.pgpEntity, downloadDestination, *signatureAsBuffer); verifyErr != nil {
					return fmt.Errorf("could not verify signature: %w", verifyErr)
				}
			}
		case "X509":
			if s.ecdsaKey == nil {
				s.relay.Debug(RlyUpdaterLog{Msg: "X509 signature provided but no local handler"})
			} else {
				//(s.pgpEntity, downloadDestination, *signatureAsBuffer)
				if verifyErr := cryptography.ECDSAVerifyFile(s.ecdsaKey, downloadDestination, s.contextUpdate.Signature); verifyErr != nil {
					return fmt.Errorf("could not verify signature: %w", verifyErr)
				}
			}
		}
	}
	s.status = updaterdto.DOWNLOADED
	return nil
}

func (s *UpdaterSvc) PerformUpdate(ctx context.Context) error {
	s.status = updaterdto.IN_PROGRESS
	if s.cfg.PrepareFunc != nil {
		updaterAgent := updaterdto.UpdaterAgentCfg{
			NetSvc:        s.netSvc,
			UpdaterCfg:    *s.cfg,
			VersionUpdate: s.contextUpdate,
		}
		if err := s.cfg.PrepareFunc(ctx, &updaterAgent); err != nil {
			return err
		}
	}

	if s.contextUpdate.ArtefactName == "" {
		return errors.New("no artefact path configured")
	}

	// Get the helper ready and validate everything is ready before proceeding
	helperPath, err := updatercopier.ExtractHelper(s.cfg.TemporaryPath)
	if err != nil {
		return err
	}
	s.relay.Info(RlyUpdaterLog{Msg: fmt.Sprintf("extracted helper to: %s", helperPath)})

	s.relay.Info(RlyUpdaterLog{Msg: fmt.Sprintf("starting update. replacing %s with %s", s.updateTarget, s.contextUpdate.ArtefactName)})
	cmd := exec.Command(helperPath, s.updateTarget, s.contextUpdate.ArtefactName, s.cfg.LogPath)
	cmd.Dir = filepath.Dir(s.cfg.TemporaryPath)
	if startErr := cmd.Start(); startErr != nil {
		return fmt.Errorf("couldn't start update helper: %w", startErr)
	}

	return nil
}
