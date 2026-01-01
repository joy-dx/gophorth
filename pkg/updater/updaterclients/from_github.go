package updaterclients

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/joy-dx/gophorth/pkg/file"
	"github.com/joy-dx/gophorth/pkg/hydrate"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/stringz"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

const UpdateClientFromGithubRef = "from_net"

type FromGithub struct {
	cfg          *FromGithubConfig
	Ref          string
	FoundVersion *releaserdto.ReleaseAsset
}

func NewFromGithub(cfg *FromGithubConfig) *FromGithub {
	return &FromGithub{
		Ref: UpdateClientFromGithubRef,
		cfg: cfg,
	}
}

func (c *FromGithub) GetRef() string {
	return c.Ref
}

func (c *FromGithub) GetVersionLink() (releaserdto.ReleaseAsset, error) {
	return *c.FoundVersion, nil
}

// TODO With both reverse template and asset name guesser, refactor combining both
func (c *FromGithub) CheckUpdate(ctx context.Context, cfg *updaterdto.UpdaterConfig) (releaserdto.ReleaseAsset, error) {
	asset := releaserdto.ReleaseAsset{}

	if hydrateErr := hydrate.NilCheck("github_check_update", map[string]interface{}{
		"client": c.cfg.Client,
		"netSvc": cfg.NetSvc,
	}); hydrateErr != nil {
		return asset, hydrateErr
	}

	var (
		foundRelease *github.RepositoryRelease
		chosenAsset  *github.ReleaseAsset
		variant      string
		err          error
	)

	if c.cfg.Tag != "" {
		foundRelease, _, err = c.cfg.Client.Repositories.GetReleaseByTag(ctx, c.cfg.Owner, c.cfg.Repo, c.cfg.Tag)
	} else {
		foundRelease, _, err = c.cfg.Client.Repositories.GetLatestRelease(ctx, c.cfg.Owner, c.cfg.Repo)
	}
	if err != nil {
		return asset, fmt.Errorf("github release fetch: %w", err)
	}

	// Pre-release handling
	if foundRelease.GetPrerelease() && !cfg.AllowPrerelease {
		return asset, fmt.Errorf("latest is prerelease (%s), but prereleases not allowed", foundRelease.GetTagName())
	}

	agentConfig := GithubAgentCfg{
		NetSvc:        cfg.NetSvc,
		UpdaterCfg:    updaterdto.UpdaterConfig{},
		GithubRelease: foundRelease,
		VersionLink:   &asset,
	}

	switch true {
	case c.cfg.SelectAssetFunc != nil:
		githubAsset, assetVariant, selErr := c.cfg.SelectAssetFunc(ctx, &agentConfig)
		if selErr != nil {
			return asset, selErr
		}
		chosenAsset, variant = githubAsset, assetVariant
	case c.cfg.SelectAssetPattern != "":
		re, compileTemplateErr := stringz.CompileReverseTemplate(stringz.ReverseTemplateOptions{
			Pattern:           c.cfg.SelectAssetPattern,
			AllowAnyExtension: true,
			RequireVersion:    false,
		})
		if compileTemplateErr != nil {
			return asset, compileTemplateErr
		}
		githubAsset := releaserdto.ReleaseAsset{
			Version: foundRelease.GetTagName(),
		}
		for idx := range foundRelease.Assets {
			name := foundRelease.Assets[idx].GetName()

			// Skip signature files that may be present
			if strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, ".asc.sig") {
				continue
			}

			matches := re.FindStringSubmatch(name)
			if matches == nil {
				continue
			}

			groupNames := re.SubexpNames()
			g := make(map[string]string, len(groupNames))
			for i, n := range groupNames {
				if i == 0 || n == "" {
					continue
				}
				g[n] = matches[i]
			}

			githubAsset.
				WithArtefactName(name).
				WithVariant(strings.TrimLeft(g["variant"], "/-_")).
				WithArch(g["arch"]).
				WithPlatform(g["platform"]).
				WithDownloadURL(foundRelease.Assets[idx].GetBrowserDownloadURL()).
				WithChecksum(foundRelease.Assets[idx].GetDigest()).
				WithSize(int64(foundRelease.Assets[idx].GetSize()))

			// Check if the asset matches our criteria
			if cfg.Platform != githubAsset.Platform ||
				cfg.Architecture != githubAsset.Arch ||
				cfg.Variant != githubAsset.Variant {
				continue
			}
			c.FoundVersion = &githubAsset
			return githubAsset, nil
		}

	default:
		// Default: best-effort choose asset based on name patterns + cfg.Platform/Architecture
		githubAsset, variantFound, selectErr := selectGitHubAssetDefault(*cfg, foundRelease.Assets)
		if selectErr != nil {
			return asset, fmt.Errorf("github asset selection: %w", selectErr)
		}
		chosenAsset, variant = githubAsset, variantFound
	}

	foundPlatform, foundArchitecture := file.AssetNameGuess(chosenAsset.GetName())
	if foundPlatform == "" || foundArchitecture == "" {
		return asset, errors.New("no asset found for platform/architecture")
	}

	asset.WithDownloadURL(chosenAsset.GetBrowserDownloadURL()).
		WithVariant(variant).
		WithPlatform(foundPlatform).
		WithArch(foundArchitecture).
		WithChecksum(chosenAsset.GetDigest()).
		WithSize(int64(chosenAsset.GetSize()))
	c.FoundVersion = &asset

	return asset, nil
}

func selectGitHubAssetDefault(cfg updaterdto.UpdaterConfig, assets []*github.ReleaseAsset) (*github.ReleaseAsset, string, error) {
	// Simple heuristic: try to match platform/arch in filename.
	wantOS := strings.ToLower(cfg.Platform)
	wantArch := strings.ToLower(cfg.Architecture)

	if wantOS != "" || wantArch != "" {
		return nil, "", errors.New("platform/arch cannot be empty")
	}

	var best *github.ReleaseAsset
	var bestVariant string

	for _, asset := range assets {
		name := strings.ToLower(asset.GetName())
		plat, arch := file.AssetNameGuess(name)
		if plat != wantOS {
			continue
		}
		if arch != wantArch {
			continue
		}

		// Prefer non-source-archives and non-checksums
		if isChecksumName(name) || isSourceArchive(name) {
			continue
		}

		// If current device wants a custom variant
		if cfg.Variant != "" {
			lowerVariant := strings.ToLower(cfg.Variant)
			if strings.Contains(name, lowerVariant) {
				bestVariant = cfg.Variant
			} else {
				continue
			}
		}

		best = asset
		break
	}

	if best == nil {
		return nil, "", errors.New("no asset found")
	}
	return best, bestVariant, nil
}

func isChecksumName(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "sha256") || strings.HasSuffix(n, ".sha256") || strings.HasSuffix(n, ".sha256sum") || strings.HasSuffix(n, ".sha256sums") || strings.HasSuffix(n, ".checksums") || strings.Contains(n, "checksums.txt")
}

func isSourceArchive(name string) bool {
	n := strings.ToLower(name)
	// Heuristic: source archives often tagged as source or src
	if strings.Contains(n, "source") || strings.Contains(n, "src") {
		ext := filepath.Ext(n)
		if ext == ".zip" || ext == ".tar" || strings.HasSuffix(n, ".tar.gz") || strings.HasSuffix(n, ".tar.xz") || strings.HasSuffix(n, ".tgz") || strings.HasSuffix(n, ".txz") {
			return true
		}
	}
	// Some projects name source tarballs with "tar.gz" and no arch/os clues; be conservative.
	return false
}
