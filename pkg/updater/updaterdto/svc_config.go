package updaterdto

import (
	"context"
	"runtime"
	"time"

	netDTO "github.com/joy-dx/gonetic/dto"
	"github.com/joy-dx/relay/dto"
)

type UpdateFuncType func(ctx context.Context, cfg *UpdaterAgentCfg) (string, error)
type PrepareFuncType func(ctx context.Context, cfg *UpdaterAgentCfg) error

// UpdaterConfig Service configuration struct
type UpdaterConfig struct {
	NetSvc netDTO.NetInterface `json:"-" yaml:"-" mapstructure:"-"`
	Relay  dto.RelayInterface  `json:"-" yaml:"-" mapstructure:"-"`
	// AllowDowngrade allows downgrading to older versions.
	AllowDowngrade bool
	// AllowPrerelease allows updating to pre-release versions.
	AllowPrerelease bool
	// Architecture If no conforming to GOOS standards, string representing architecture part
	Architecture string `json:"architecture" yaml:"architecture" mapstructure:"architecture"`
	// Platform If no conforming to GOOS standards, string representing platform part
	Platform string `json:"platform" yaml:"platform" mapstructure:"platform"`
	// PublicKey Contains ASCII encode EDCSA or PGP public key
	PublicKey string `json:"public_key" yaml:"public_key" mapstructure:"public_key"`
	// PublicKeyPath Path to EDCSA or PGP public key
	PublicKeyPath string `json:"public_key_path" yaml:"public_key_path" mapstructure:"public_key_path"`
	// TemporaryPath Where to store download and update artefacts
	TemporaryPath string `json:"temporary_path" yaml:"temporary_path" mapstructure:"temporary_path"`
	// Variant Represents a download variant that the current device wants
	Variant string `json:"variant" yaml:"variant" mapstructure:"variant"`
	// CheckInterval Adding support for periodic checks
	CheckInterval time.Duration `json:"check_interval,omitempty" yaml:"check_interval,omitempty" mapstructure:"check_interval"`
	// Version Semantic version representing current runtime version
	Version string `json:"version" yaml:"version" mapstructure:"version"`
	// LastUpdateCheck Represents the last lookup in Go time
	LastUpdateCheck *time.Time `json:"last_update_check,omitempty" yaml:"last_update_check,omitempty" mapstructure:"last_update_check"`
	// LogPath Local file system path used during update as log path
	LogPath string `json:"log_path,omitempty" yaml:"log_path,omitempty" mapstructure:"log_path"`
	// CheckClient Agent for retrieving update information
	CheckClient CheckClientInterface `json:"-" yaml:"-" mapstructure:"-"`
	// DownloadFunc Optional override for downloading the artefact
	DownloadFunc UpdateFuncType `json:"-" yaml:"-" mapstructure:"-"`
	// PrepareFunc Preupdate preparation returning path for update material
	PrepareFunc PrepareFuncType `json:"-" yaml:"-" mapstructure:"-"`
	// Verifiers Additional procedures for verifying update integrity
	Verifiers []VerificationMethodInterface `json:"-" yaml:"-" mapstructure:"-"`
}

func DefaultUpdaterSvcConfig() UpdaterConfig {
	return UpdaterConfig{
		LogPath:       ".",
		CheckInterval: 48 * time.Hour,
		Architecture:  runtime.GOARCH,
		Platform:      runtime.GOOS,
		TemporaryPath: "/tmp/gophorth",
	}
}

func (c *UpdaterConfig) WithArch(arch string) *UpdaterConfig {
	c.Architecture = arch
	return c
}

func (c *UpdaterConfig) WithLastUpdateCheck(time *time.Time) *UpdaterConfig {
	c.LastUpdateCheck = time
	return c
}

func (c *UpdaterConfig) WithPlatform(platform string) *UpdaterConfig {
	c.Platform = platform
	return c
}

func (c *UpdaterConfig) WithUpdateCheckInterval(interval time.Duration) *UpdaterConfig {
	c.CheckInterval = interval
	return c
}

func (c *UpdaterConfig) WithUpdateLogPath(path string) *UpdaterConfig {
	c.LogPath = path
	return c
}

func (c *UpdaterConfig) WithCheckClient(client CheckClientInterface) *UpdaterConfig {
	c.CheckClient = client
	return c
}

func (c *UpdaterConfig) WithVersion(version string) *UpdaterConfig {
	c.Version = version
	return c
}

func (c *UpdaterConfig) WithPrepareFunc(userFunc PrepareFuncType) *UpdaterConfig {
	c.PrepareFunc = userFunc
	return c
}

func (c *UpdaterConfig) WithPublicKey(key string) *UpdaterConfig {
	c.PublicKey = key
	return c
}

func (c *UpdaterConfig) WithPublicKeyPath(path string) *UpdaterConfig {
	c.PublicKeyPath = path
	return c
}

func (c *UpdaterConfig) WithTemporaryPath(path string) *UpdaterConfig {
	c.TemporaryPath = path
	return c
}

func (c *UpdaterConfig) WithUpdateFunc(userFunc UpdateFuncType) *UpdaterConfig {
	c.DownloadFunc = userFunc
	return c
}

func (c *UpdaterConfig) WithVariant(variant string) *UpdaterConfig {
	c.Variant = variant
	return c
}

func (c *UpdaterConfig) WithVerifier(client VerificationMethodInterface) *UpdaterConfig {
	c.Verifiers = append(c.Verifiers, client)
	return c
}

func (c *UpdaterConfig) WithVerifiers(client []VerificationMethodInterface) *UpdaterConfig {
	c.Verifiers = client
	return c
}

func (c *UpdaterConfig) WithNetSvc(svc netDTO.NetInterface) *UpdaterConfig {
	c.NetSvc = svc
	return c
}

func (c *UpdaterConfig) WithRelay(relay dto.RelayInterface) *UpdaterConfig {
	c.Relay = relay
	return c
}
