package updaterdto

import (
	"time"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

type UpdaterState struct {
	LastTimeCheckedUpdate *time.Time                `json:"updater_last_time_checked_update" ts_type:"string"`
	UpdateLink            *releaserdto.ReleaseAsset `json:"updater_update_link"`
	Changelog             string                    `json:"updater_changelog"`
	ReleasedAt            *time.Time                `json:"updater_released_at" ts_type:"string"`
	CheckInterval         time.Duration             `json:"updater_check_interval"`
	Log                   string                    `json:"updater_log"`
	LogPath               string                    `json:"updater_log_path"`
	Updating              bool                      `json:"updater_updating"`
	Version               string                    `json:"updater_version" yaml:"updater_version"`
}

type UpdaterAgentCfg struct {
	NetSvc        netdto.NetInterface
	UpdaterCfg    UpdaterConfig
	VersionUpdate *releaserdto.ReleaseAsset
}
