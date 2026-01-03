package updaterclients

import (
	"context"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay/dto"
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
	NetSvc      netdto.NetInterface
	Relay       dto.RelayInterface
	UpdaterCfg  updaterdto.UpdaterConfig
	VersionLink *releaserdto.ReleaseAsset
}
