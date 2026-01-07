package updaterclients

import (
	"context"

	"github.com/google/go-github/v81/github"
	netDTO "github.com/joy-dx/gonetic/dto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

type SelectAssetFuncType func(ctx context.Context, cfg *GithubAgentCfg) (*github.ReleaseAsset, string /*variant*/, error)
type GetSignatureFuncType func(ctx context.Context, cfg *GithubAgentCfg) (string, error)

// FromGithubConfig Service configuration struct
type FromGithubConfig struct {
	// Pre-initialized GitHub client (optionally with auth)
	Client *github.Client
	Owner  string
	Repo   string
	// Optional: if specified, use that tag; otherwise, latest release.
	Tag string
	// SelectAssetPattern A template style string representing the file name wanted
	SelectAssetPattern string `json:"select_asset_pattern"`
	// Optional asset filter callback; if nil, AssetNameGuess + cfg filters apply.
	SelectAssetFunc SelectAssetFuncType
	// Optional to download key ready for later verification
	GetSignatureFunc GetSignatureFuncType
}

func DefaultFromGithubConfig() FromGithubConfig {
	return FromGithubConfig{}
}

func (c *FromGithubConfig) GetRef() string {
	return UpdateClientFromGithubRef + "_config"
}

func (c *FromGithubConfig) WithClient(client *github.Client) *FromGithubConfig {
	c.Client = client
	return c
}

func (c *FromGithubConfig) WithOwner(owner string) *FromGithubConfig {
	c.Owner = owner
	return c
}

func (c *FromGithubConfig) WithRepo(repo string) *FromGithubConfig {
	c.Repo = repo
	return c
}

func (c *FromGithubConfig) WithTag(tag string) *FromGithubConfig {
	c.Tag = tag
	return c
}

func (c *FromGithubConfig) WithSelectAssetPattern(pattern string) *FromGithubConfig {
	c.SelectAssetPattern = pattern
	return c
}

func (c *FromGithubConfig) WithSelectAssetFunc(userFunc SelectAssetFuncType) *FromGithubConfig {
	c.SelectAssetFunc = userFunc
	return c
}

func (c *FromGithubConfig) WithGetSignatureFunc(userFunc GetSignatureFuncType) *FromGithubConfig {
	c.GetSignatureFunc = userFunc
	return c
}

type GithubAgentCfg struct {
	NetSvc        netDTO.NetInterface
	UpdaterCfg    updaterdto.UpdaterConfig
	GithubRelease *github.RepositoryRelease
	VersionLink   *releaserdto.ReleaseAsset
}
