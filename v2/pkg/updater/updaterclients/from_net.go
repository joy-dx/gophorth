package updaterclients

import (
	"context"
	"errors"

	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

const UpdateClientFromNetRef = "from_net"

type FromNet struct {
	cfg          *FromNetConfig
	Ref          string
	FoundVersion releaserdto.ReleaseAsset
}

func NewFromNet(cfg *FromNetConfig) *FromNet {
	return &FromNet{
		Ref: UpdateClientFromNetRef,
		cfg: cfg,
	}
}

func (c *FromNet) GetRef() string {
	return c.Ref
}

func (c *FromNet) GetVersionLink() (releaserdto.ReleaseAsset, error) {
	return c.FoundVersion, nil
}

func (c *FromNet) CheckUpdate(ctx context.Context, cfg *updaterdto.UpdaterConfig) (releaserdto.ReleaseAsset, error) {
	if c.cfg.UserFetchFunction == nil {
		return releaserdto.ReleaseAsset{}, errors.New("FromNetCheckClient: missing UserFetchFunction")
	}
	agentConfig := NetAgentCfg{
		NetSvc:     cfg.NetSvc,
		Relay:      cfg.Relay,
		UpdaterCfg: *cfg,
	}
	versionLink, err := c.cfg.UserFetchFunction(ctx, agentConfig)
	if err != nil {
		return releaserdto.ReleaseAsset{}, err
	}
	c.FoundVersion = versionLink
	return c.FoundVersion, nil
}
