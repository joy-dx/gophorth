package relayconfig

import "github.com/joy-dx/gophorth/pkg/relay/relaydto"

type RelaySvcConfig struct {
	Sinks []relaydto.RelaySinkInterface `json:"sinks" yaml:"sinks" mapstructure:"sinks"`
}

func DefaultRelaySvcConfig() RelaySvcConfig {
	return RelaySvcConfig{}
}
