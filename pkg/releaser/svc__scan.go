package releaser

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joy-dx/gophorth/pkg/cryptography"
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

	re, err := s.compileReverseTemplate()
	if err != nil {
		return nil, err
	}

	s.relay.Info(RlyReleaserLog{Msg: fmt.Sprintf("starting scan: %s", s.cfg.TargetPath)})
	targetPath := os.ExpandEnv(s.cfg.TargetPath)
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", s.cfg.TargetPath, err)
	}

	out := make([]releaserdto.ReleaseAsset, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		// Skip signature files that may be present
		if strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, ".asc.sig") {
			continue
		}

		matches := re.FindStringSubmatch(name)
		if matches == nil {
			if s.cfg.Strict {
				return nil, fmt.Errorf("file %q does not match pattern %q", name, s.cfg.FilePattern)
			}
			continue
		}

		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", name, err)
		}

		groupNames := re.SubexpNames()
		g := make(map[string]string, len(groupNames))
		for i, n := range groupNames {
			if i == 0 || n == "" {
				continue
			}
			g[n] = matches[i]
		}

		fullPath := filepath.Join(s.cfg.TargetPath, name)

		checksum, checksumErr := cryptography.Sha256SumFile(fullPath)
		if checksumErr != nil {
			return nil, fmt.Errorf("checksum file %q: %w", fullPath, checksumErr)
		}

		s.checksumBuilder.WriteString(fmt.Sprintf("%s  %s\n", checksum, path.Base(fullPath)))
		var version string
		if s.version != nil {
			version = s.version.String()
		}
		foundVersion := trimLeadingDash(g["version"])
		if foundVersion != "" {
			version = foundVersion
		}
		out = append(out, releaserdto.ReleaseAsset{
			ArtefactName: filepath.Base(fullPath),
			Platform:     g["platform"],
			Arch:         g["arch"],
			Variant:      g["variant"],
			Version:      version,
			SizeBytes:    info.Size(),
			Checksum:     checksum,
		})
	}
	s.releaseAssets = out
	return out, nil
}

func trimLeadingDash(s string) string {
	if strings.HasPrefix(s, "-") {
		return s[1:]
	}
	return s
}

// compileReverseTemplate turns a template string into a regex with named groups.
// Supported placeholders: {platform}, {arch}, {variant}, {version}.
//   - {variant} is optional and includes its leading "-" when present
//     (e.g. "-webkit241"). Captured as "-webkit241" and later kept as-is.
//   - {version} includes its leading "-" when present (e.g. "-1.2.3").
//     If RequireVersion is false, it's optional; if true, it's required.
//
// Example template: "test-app-{platform}-{arch}{variant}{version}"
func (s *ReleaserSvc) compileReverseTemplate() (*regexp.Regexp, error) {
	if strings.TrimSpace(s.cfg.FilePattern) == "" {
		return nil, errors.New("pattern must not be empty")
	}

	var versionRule string
	if s.cfg.RequireVersion {
		versionRule = `-[0-9]+(?:\.[0-9A-Za-z]+)*(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?`
	} else {
		versionRule = `(?:-[0-9]+(?:\.[0-9A-Za-z]+)*(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?)?`
	}

	rules := map[string]string{
		"platform": `[a-z0-9]+`,
		"arch":     `[A-Za-z0-9_]+`,
		"variant":  `(?:-[A-Za-z0-9._]+)?`,
		"version":  versionRule,
	}

	var b strings.Builder
	b.WriteString("^")

	for i := 0; i < len(s.cfg.FilePattern); {
		if s.cfg.FilePattern[i] != '{' {
			j := strings.IndexByte(s.cfg.FilePattern[i:], '{')
			if j == -1 {
				j = len(s.cfg.FilePattern)
			} else {
				j = i + j
			}
			b.WriteString(regexp.QuoteMeta(s.cfg.FilePattern[i:j]))
			i = j
			continue
		}

		end := strings.IndexByte(s.cfg.FilePattern[i:], '}')
		if end == -1 {
			return nil, fmt.Errorf("unclosed placeholder starting at index %d", i)
		}
		end = i + end

		name := s.cfg.FilePattern[i+1 : end]
		if name == "" {
			return nil, fmt.Errorf("empty placeholder at index %d", i)
		}

		rule, ok := rules[name]
		if !ok {
			return nil, fmt.Errorf("unknown placeholder {%s}", name)
		}

		b.WriteString("(?P<")
		b.WriteString(name)
		b.WriteString(">")
		b.WriteString(rule)
		b.WriteString(")")

		i = end + 1
	}

	if s.cfg.AllowAnyExtension {
		// Allow ".zip", ".tar.gz", etc. after the pattern.
		// .asc/.asc.sig are filtered by caller since Go regex lacks lookarounds.
		b.WriteString(`(?:\.[A-Za-z0-9]+(?:\.[A-Za-z0-9]+)*)?`)
	}

	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil, fmt.Errorf("compile regex from pattern %q: %w", s.cfg.FilePattern, err)
	}
	return re, nil
}
