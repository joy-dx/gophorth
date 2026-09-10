package updaterclients

import (
	"context"

	netDTO "github.com/joy-dx/gonetic/v2/dto"
	"github.com/joy-dx/gophorth/v2/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/v2/pkg/updater/updaterdto"
	"github.com/joy-dx/relay/v2/dto"
)

type NetCheckFunc func(ctx context.Context, config NetAgentCfg) (releaserdto.ReleaseAsset, error)

// FromGithubConfig Service configuration struct
type FromNetConfig struct {
	UserFetchFunction NetCheckFunc
}

func DefaultFromNetConfig() FromNetConfig {
	return FromNetConfig{}
}

func (c *FromNetConfig) GetRef() string {
	return UpdateClientFromNetRef + "_config"
}

func (c *FromNetConfig) WithUserFetchFunction(checkFunc NetCheckFunc) *FromNetConfig {
	c.UserFetchFunction = checkFunc
	return c
}

type NetAgentCfg struct {
	NetSvc      netDTO.NetInterface
	Relay       dto.RelayInterface
	UpdaterCfg  updaterdto.UpdaterConfig
	VersionLink *releaserdto.ReleaseAsset
}
