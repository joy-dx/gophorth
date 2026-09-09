package releaser

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/file/parser"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

// ScanDir reads a directory (non-recursive) and returns ReleasesFound entries
// parsed from filenames using a pattern like:
//
//	"test-app-{platform}-{arch}{variant}{version}"
//
// Examples:
//
//	"test-app-linux-amd64-webkit241-1.2.3.zip"
//	"test-app-darwin-arm64.zip"
//
// Behavior:
//   - {variant} is optional and, when present, includes its leading dash (e.g. "-webkit241").
//   - {version} is optional by default and, when present, includes its leading dash
//     (e.g. "-1.2.3"). Set RequireVersion=true to make it required.
func (s *ReleaserSvc) ScanDir() ([]releaserdto.ReleaseAsset, error) {
	s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("using pattern: %s", s.cfg.FilePattern)})
	compiledParser, err := parser.CompileFilenamePattern(s.cfg.FilePattern)
	if err != nil {
		return nil, fmt.Errorf("scan dir. failed to compile pattern. %s. %w", s.cfg.FilePattern, err)
	}

	s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("starting scan: %s", s.cfg.TargetPath)})
	targetPath := os.ExpandEnv(s.cfg.TargetPath)
	dirListing, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", s.cfg.TargetPath, err)
	}

	out := make([]releaserdto.ReleaseAsset, 0, len(dirListing))
	s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("found %d possible assets", len(dirListing))})
	for _, dirEntry := range dirListing {
		if dirEntry.IsDir() {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("skipping dir: %s", dirEntry.Name())})
			continue
		}

		// Skip signature files that may be present
		if strings.HasSuffix(dirEntry.Name(), ".asc") || strings.HasSuffix(dirEntry.Name(), ".sig") {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("skipping signature: %s", dirEntry.Name())})
			continue
		}

		fileMeta, parseErr := compiledParser.ParseFilename(dirEntry.Name())
		if parseErr != nil {
			s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to parse filename: %s. %s", dirEntry.Name(), parseErr.Error())})
			continue
		}

		fullPath := filepath.Join(s.cfg.TargetPath, dirEntry.Name())

		statInfo, statErr := os.Stat(fullPath)
		if statErr != nil {
			s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to stat: %s. %s", dirEntry.Name(), statErr.Error())})
			continue
		}

		checksum, checksumErr := cryptography.Sha256SumFile(fullPath)
		if checksumErr != nil {
			s.relay.Warn(RlyReleaserLog{Msg: fmt.Sprintf("failed to generate checksum: %s. %s", dirEntry.Name(), checksumErr.Error())})
			continue
		}

		s.checksumBuilder.WriteString(fmt.Sprintf("%s  %s\n", checksum, path.Base(fullPath)))
		var version string

		if fileMeta.Artifact.Version != "" {
			s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("found version: %s", fileMeta.Artifact.Version)})
			version = fileMeta.Artifact.Version
		} else {
			if s.version != nil {
				s.relay.Debug(RlyReleaserLog{Msg: fmt.Sprintf("using version from service: %s", s.version.String())})
				version = s.version.String()
			}
		}

		out = append(out, releaserdto.ReleaseAsset{
			ArtefactName: filepath.Base(fullPath),
			Platform:     fileMeta.Artifact.Platform,
			Arch:         fileMeta.Artifact.Arch,
			Variant:      fileMeta.Artifact.Variant,
			Version:      version,
			SizeBytes:    statInfo.Size(),
			Checksum:     checksum,
		})
	}
	s.releaseAssets = out
	return out, nil
}
