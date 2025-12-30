package releaserdto

import (
	"context"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

type ProcessReleasesFuncType func(ctx context.Context, config AgentCfg) error

// ReleaserConfig Service configuration struct
type ReleaserConfig struct {
	NetSvc         netdto.NetInterface
	Relay          relaydto.RelayInterface
	DownloadPrefix string `json:"download_prefix" yaml:"download_prefix"`
	// OutputPath FS Path where generated artefacts will be saved
	OutputPath          string `json:"output_path" yaml:"output_path" mapstructure:"output_path"`
	ProcessReleasesFunc ProcessReleasesFuncType
	// TargetPath FS path to published artefacts
	TargetPath string `json:"target_path" yaml:"target_path" mapstructure:"target_path"`
	// FilePattern name the published app to be processed starts with
	FilePattern        string `json:"file_pattern" yaml:"file_pattern" mapstructure:"file_pattern"`
	GenerateChecksums  bool   `json:"generate_checksums" yaml:"generate_checksums" mapstructure:"generate_checksums"`
	GenerateSignatures bool   `json:"generate_signatures" yaml:"generate_signatures" mapstructure:"generate_signatures"`
	// PrivateKey Contains ASCII encode EDCSA or PGP public key
	PrivateKey string `json:"private_key" yaml:"private_key" mapstructure:"private_key"`
	// PrivateKeyPath Path to EDCSA or PGP public key
	PrivateKeyPath string `json:"private_key_path" yaml:"private_key_path" mapstructure:"private_key_path"`
	// SummaryOutputType Controls the file format for export
	SummaryOutputType string `json:"summary_output_type" yaml:"summary_output_type" mapstructure:"summary_output_type"`
	// Allows a file extension after the pattern (".zip", ".tar.gz", etc.)
	AllowAnyExtension bool `json:"allow_any_extension" yaml:"allow_any_extension" mapstructure:"allow_any_extension"`
	// If true, non-matching files cause an error. If false, they are skipped.
	Strict bool `json:"strict" yaml:"strict" mapstructure:"strict"`
	// If true, {version} is treated as required when used in the pattern.
	// If false, {version} is optional (the capture is "", not present).
	RequireVersion bool `json:"require_version" yaml:"require_version" mapstructure:"require_version"`
	// Version Manually specify version to use with release
	Version string `json:"version" yaml:"version" mapstructure:"version"`
}

func DefaultReleaserConfig() ReleaserConfig {
	return ReleaserConfig{
		GenerateChecksums:  true,
		GenerateSignatures: true,
		SummaryOutputType:  "json-indented",
	}
}

func (c *ReleaserConfig) WithAllowAnyExtension(truthy bool) *ReleaserConfig {
	c.AllowAnyExtension = truthy
	return c
}

func (c *ReleaserConfig) WithDownloadPrefix(prefix string) *ReleaserConfig {
	c.DownloadPrefix = prefix
	return c
}

func (c *ReleaserConfig) WithGenerateChecksums(truthy bool) *ReleaserConfig {
	c.GenerateChecksums = truthy
	return c
}

func (c *ReleaserConfig) WithGenerateSignatures(truthy bool) *ReleaserConfig {
	c.GenerateSignatures = truthy
	return c
}

func (c *ReleaserConfig) WithOutputPath(path string) *ReleaserConfig {
	c.OutputPath = path
	return c
}

func (c *ReleaserConfig) WithRequireVersion(truthy bool) *ReleaserConfig {
	c.RequireVersion = truthy
	return c
}

func (c *ReleaserConfig) WithStrict(truthy bool) *ReleaserConfig {
	c.Strict = truthy
	return c
}

func (c *ReleaserConfig) WithTargetPath(path string) *ReleaserConfig {
	c.TargetPath = path
	return c
}

func (c *ReleaserConfig) WithFilePattern(pattern string) *ReleaserConfig {
	c.FilePattern = pattern
	return c
}

func (c *ReleaserConfig) WithPrivateKey(key string) *ReleaserConfig {
	c.PrivateKey = key
	return c
}

func (c *ReleaserConfig) WithPrivateKeyPath(path string) *ReleaserConfig {
	c.PrivateKeyPath = path
	return c
}

func (c *ReleaserConfig) WithProcessReleasesFunc(userFunc ProcessReleasesFuncType) *ReleaserConfig {
	c.ProcessReleasesFunc = userFunc
	return c
}

func (c *ReleaserConfig) WithSummaryOutputType(outputType string) *ReleaserConfig {
	c.SummaryOutputType = outputType
	return c
}

func (c *ReleaserConfig) WithVersion(version string) *ReleaserConfig {
	c.Version = version
	return c
}

func (c *ReleaserConfig) WithNetSvc(svc netdto.NetInterface) *ReleaserConfig {
	c.NetSvc = svc
	return c
}

func (c *ReleaserConfig) WithRelay(relay relaydto.RelayInterface) *ReleaserConfig {
	c.Relay = relay
	return c
}
