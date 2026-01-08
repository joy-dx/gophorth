package updaterdto

import (
	"time"

	netDTO "github.com/joy-dx/gonetic/dto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

type UpdaterState struct {
	Architecture    string                    `json:"updater_architecture"`
	Changelog       string                    `json:"updater_changelog"`
	CheckInterval   time.Duration             `json:"updater_check_interval"`
	LastUpdateCheck *time.Time                `json:"updater_last_update_check" ts_type:"string"`
	Log             string                    `json:"updater_log"`
	LogPath         string                    `json:"updater_log_path"`
	Platform        string                    `json:"updater_platform"`
	PublicKey       string                    `json:"updater_public_key"`
	PublicKeyPath   string                    `json:"updater_public_key_path"`
	ReleasedAt      *time.Time                `json:"updater_released_at" ts_type:"string"`
	TemporaryPath   string                    `json:"updater_temporary_path"`
	UpdateLink      *releaserdto.ReleaseAsset `json:"updater_update_link"`
	Updating        bool                      `json:"updater_updating"`
	Variant         string                    `json:"updater_variant"`
	Version         string                    `json:"updater_version" yaml:"updater_version"`
}

type UpdaterAgentCfg struct {
	NetSvc        netDTO.NetInterface
	UpdaterCfg    UpdaterConfig
	VersionUpdate *releaserdto.ReleaseAsset
}
