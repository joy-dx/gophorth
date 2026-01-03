package releaser

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/file"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/relay/dto"
)

type ReleaserSvc struct {
	relay               dto.RelayInterface
	cfg                 *releaserdto.ReleaserConfig
	changelog           string
	checksumBuilder     strings.Builder
	binarySigningMethod string
	pgpEntity           openpgp.EntityList
	ecdsaKey            *ecdsa.PrivateKey
	releasedAt          *time.Time
	version             *semver.Version
	releaseAssets       []releaserdto.ReleaseAsset
}

func (s *ReleaserSvc) GenerateArtefacts(ctx context.Context) (releaserdto.ReleaseAsset, error) {
	var update releaserdto.ReleaseAsset

	return update, nil
}

func (s *ReleaserSvc) SignFiles(ctx context.Context) error {
	return nil
}

func (s *ReleaserSvc) GenerateReleaseSummary(ctx context.Context) (releaserdto.ReleaseSummary, error) {

	releasesFound, err := s.ScanDir()
	if err != nil {
		return releaserdto.ReleaseSummary{}, fmt.Errorf("problem scanning releases: %w", err)
	}

	numberOfReleases := len(releasesFound)
	s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("found %d releases", numberOfReleases)})

	if numberOfReleases == 0 {
		return releaserdto.ReleaseSummary{}, errors.New("no releases found")
	}

	if s.cfg.DownloadPrefix != "" {
		for idx := range releasesFound {
			releasesFound[idx].DownloadURL = s.cfg.DownloadPrefix + releasesFound[idx].ArtefactName
		}
	}

	if s.cfg.GenerateSignatures && s.binarySigningMethod != "" {
		s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("signing releases by: %s", s.binarySigningMethod)})
		switch s.binarySigningMethod {
		case "PGP":
			for idx, release := range releasesFound {
				artefactPath := s.cfg.TargetPath + "/" + release.ArtefactName

				signature, sigErr := cryptography.PGPSignFile(s.pgpEntity, artefactPath)
				if sigErr != nil {
					s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to sign %s: %s", artefactPath, sigErr.Error())})
				}
				release.SignatureType = "PGP"
				release.Signature = signature

				signatureTargetPath := os.ExpandEnv(s.cfg.OutputPath + "/" + release.ArtefactName + ".asc")

				if writeErr := file.BytesToFile([]byte(signature), signatureTargetPath); writeErr != nil {
					s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to write signature: %s", writeErr.Error())})
				}

				releasesFound[idx] = release
			}
		case "X509":
			for idx, release := range releasesFound {
				artefactPath := s.cfg.TargetPath + "/" + release.ArtefactName
				signature, sigErr := cryptography.ECDSASignFile(s.ecdsaKey, artefactPath)
				if sigErr != nil {
					s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to sign %s: %s", artefactPath, sigErr.Error())})
				}
				release.SignatureType = "X509"
				release.Signature = signature
				signatureTargetPath := os.ExpandEnv(s.cfg.OutputPath + "/" + release.ArtefactName + ".asc")

				if writeErr := file.BytesToFile([]byte(signature), signatureTargetPath); writeErr != nil {
					s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to write signature: %s", writeErr.Error())})
				}

				releasesFound[idx] = release
			}
		}
	}
	s.releaseAssets = releasesFound

	if s.cfg.GenerateChecksums {
		checksumTargetPath := os.ExpandEnv(s.cfg.OutputPath + "/checksums.txt")
		s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("writing checksums to: %s", checksumTargetPath)})
		if writeErr := file.BytesToFile([]byte(s.checksumBuilder.String()), checksumTargetPath); writeErr != nil {
			return releaserdto.ReleaseSummary{}, fmt.Errorf("problem writing checksum: %w", writeErr)
		}
	}

	if s.cfg.ProcessReleasesFunc != nil {
		agentCfg := releaserdto.AgentCfg{
			NetSvc:        s.cfg.NetSvc,
			UpdaterCfg:    *s.cfg,
			ReleasesFound: s.releaseAssets,
		}
		if processErr := s.cfg.ProcessReleasesFunc(ctx, agentCfg); processErr != nil {
			s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("Failed to process releases for release %s: %s", s.cfg.NetSvc, processErr.Error())})
		}
	}

	now := time.Now()
	releaseSummary := releaserdto.ReleaseSummary{
		Assets:      releasesFound,
		PublishedAt: &now,
		Version:     s.cfg.Version,
	}

	s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("outputting release summary as %s to: %s", s.cfg.SummaryOutputType, os.ExpandEnv(s.cfg.OutputPath))})
	switch s.cfg.SummaryOutputType {
	case "json":
		outputFilePath := os.ExpandEnv(s.cfg.OutputPath + "/version.json")
		if writeErr := file.StructToJSONFile(releaseSummary, outputFilePath); writeErr != nil {
			return releaserdto.ReleaseSummary{}, fmt.Errorf("problem writing summary: %w", writeErr)
		}
	case "json-indented":
		outputFilePath := os.ExpandEnv(s.cfg.OutputPath + "/version.json")
		if writeErr := file.StructToIndentedJSONFile(releaseSummary, outputFilePath); writeErr != nil {
			return releaserdto.ReleaseSummary{}, fmt.Errorf("problem writing summary: %w", writeErr)
		}
	case "yaml":
		outputFilePath := os.ExpandEnv(s.cfg.OutputPath + "/version.yaml")
		if writeErr := file.StructToYamlFile(releaseSummary, outputFilePath); writeErr != nil {
			return releaserdto.ReleaseSummary{}, fmt.Errorf("problem writing summary: %w", writeErr)
		}
	}

	return releaseSummary, nil
}
