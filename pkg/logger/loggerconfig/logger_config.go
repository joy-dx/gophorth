package loggerconfig

import "github.com/joy-dx/gophorth/pkg/relay/relaydto"

type LoggerConfig struct {
	// Level What structured level of log to cutoff display
	Level relaydto.RelayLevel `json:"level" yaml:"level" mapstructure:"level"`
	// Type Whether to use the simple message or structured log handler
	Type       string `json:"type" yaml:"type" mapstructure:"type"`
	KeyPadding int    `json:"-,omitempty" yaml:"-,omitempty" mapstructure:"key_padding,omitempty"`
}

func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Type:  "structured",
		Level: relaydto.Debug,
	}
}

func (c *LoggerConfig) WithKeyPadding(padding int) *LoggerConfig {
	c.KeyPadding = padding
	return c
}

func (c *LoggerConfig) WithLevel(level relaydto.RelayLevel) *LoggerConfig {
	c.Level = level
	return c
}

func (c *LoggerConfig) WithType(typeName string) *LoggerConfig {
	c.Type = typeName
	return c
}
