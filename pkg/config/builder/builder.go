package builder

import (
	"log"
	"strings"
	"time"

	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO Hide config parameter

type ConfigBuilder struct {
	cmd          *cobra.Command
	configPrefix []string
}

func (b *ConfigBuilder) SetCommand(cmd *cobra.Command) {
	b.cmd = cmd
}

func (b *ConfigBuilder) SetConfigPrefix(prefix []string) {
	b.configPrefix = prefix
}

func CobraKey(prefix []string, configKey options.ConfigOption) string {
	var sb strings.Builder
	for _, prefixPiece := range prefix {
		sb.WriteString(prefixPiece)
		sb.WriteString("-")
	}
	sb.WriteString(strings.ReplaceAll(string(configKey), "_", "-"))
	return sb.String()
}

func (b *ConfigBuilder) cobraParamName(configKey options.ConfigOption) string {
	return CobraKey(b.configPrefix, configKey)
}

func ViperKey(prefix []string, configKey options.ConfigOption) string {
	var sb strings.Builder
	for _, prefixPiece := range prefix {
		sb.WriteString(prefixPiece)
		sb.WriteString(".")
	}
	sb.WriteString(string(configKey))
	return sb.String()
}

func (b *ConfigBuilder) viperParamName(configKey options.ConfigOption) string {
	return ViperKey(b.configPrefix, configKey)
}

func (b *ConfigBuilder) AddBoolParam(configKey options.ConfigOption, defaultValue bool, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().Bool(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddDurationParam(configKey options.ConfigOption, defaultValue time.Duration, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().Duration(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddIntParam(configKey options.ConfigOption, defaultValue int, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().Int(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddStringParam(configKey options.ConfigOption, defaultValue string, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().String(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddStringHiddenParam(configKey options.ConfigOption, defaultValue string, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().String(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddStringMapParam(configKey options.ConfigOption, defaultValue map[string]string, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().StringToString(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}

func (b *ConfigBuilder) AddStringSliceParam(configKey options.ConfigOption, defaultValue []string, description string) *ConfigBuilder {
	cobraParam := b.cobraParamName(configKey)
	viperParam := b.viperParamName(configKey)
	b.cmd.PersistentFlags().StringSlice(cobraParam, defaultValue, description)
	if err := viper.BindPFlag(viperParam, b.cmd.PersistentFlags().Lookup(cobraParam)); err != nil {
		log.Fatal(err)
	}
	return b
}
