package releaserdto

import (
	"time"

	netDTO "github.com/joy-dx/gonetic/dto"
)

type ReleaserState struct {
	Assets     []ReleaseAsset `json:"releaser_assets"`
	Changelog  string         `json:"releaser_changelog"`
	ReleasedAt *time.Time     `json:"releaser_released_at" ts_type:"string"`
}

type AgentCfg struct {
	NetSvc        netDTO.NetInterface
	UpdaterCfg    ReleaserConfig
	ReleasesFound []ReleaseAsset
}

// ReleaseSummary Represents
type ReleaseSummary struct {
	Changelog   string         `json:"changelog,omitempty"`
	Assets      []ReleaseAsset `json:"assets"`
	PublishedAt *time.Time     `json:"published_at"`
	ReleaseURL  string         `json:"release_url"`
	Version     string         `json:"version"`
}
