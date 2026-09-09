package releaser

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/file"
	"github.com/joy-dx/gophorth/pkg/hydrate"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

func (s ReleaserSvc) State() *releaserdto.ReleaserState {
	return &releaserdto.ReleaserState{
		Assets:     s.releaseAssets,
		Changelog:  s.changelog,
		ReleasedAt: s.releasedAt,
	}
}

func (s *ReleaserSvc) Hydrate(ctx context.Context) error {
	if err := hydrate.NilCheck("net", map[string]interface{}{
		"config": s.cfg,
		"relay":  s.relay,
	}); err != nil {
		return err
	}

	if s.cfg.PrivateKeyPath != "" {
		s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("Loading private key from %s", s.cfg.PrivateKeyPath)})

		privateKey, err := file.ToBytes(s.cfg.PrivateKeyPath)
		if err != nil {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("failed to read private key file %s", s.cfg.PrivateKeyPath)})
		}
		s.cfg.WithPrivateKey(string(privateKey))
	}
	if s.cfg.PrivateKey != "" {
		if keyInfo, err := cryptography.DetectSignatureInformation([]byte(s.cfg.PrivateKey)); err != nil {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("could not detect key information: %s", err.Error())})
		} else {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("using key format: %s", keyInfo.Format)})
			switch keyInfo.Format {
			case "PGP":
				pgpEntity, keyringErr := cryptography.LoadKeyRingAuto([]byte(s.cfg.PrivateKey))
				if keyringErr != nil {
					s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("could not load public key information: %s", keyringErr.Error())})
				}
				s.pgpEntity = pgpEntity
				s.binarySigningMethod = "PGP"
			case "X509":
				ecdsaKey, keyringErr := cryptography.ParseECDSAPrivateKeyFromPEM(s.cfg.PrivateKey)
				if keyringErr != nil {
					s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("could not load public key information: %s", keyringErr.Error())})
				}
				s.ecdsaKey = ecdsaKey
				s.binarySigningMethod = "X509"
			default:
				s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("unsupported key format: %s", keyInfo.Format)})
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
	s.relay.Debug(RlyReleaserLog{Msg: "end: hydrate state"})
	return nil
}
