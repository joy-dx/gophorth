package config

import (
	"os"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	netCfg "github.com/joy-dx/gonetic/config"
	"github.com/joy-dx/gophorth/pkg/file"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay/config"
	"github.com/spf13/viper"
)

// ConfigSvc Handles configuration parameters during program initiation. Only needed if using flags or saving config to file.
type ConfigSvc struct {
	cfgFilePath          string
	cfgEnvironmentPrefix string
	Net                  netCfg.NetSvcConfig        `json:"net" yaml:"net" mapstructure:"net"`
	Relay                config.RelaySvcConfig      `json:"relay" yaml:"relay" mapstructure:"relay"`
	Releaser             releaserdto.ReleaserConfig `json:"releaser" yaml:"releaser" mapstructure:"releaser"`
	Updater              updaterdto.UpdaterConfig   `json:"updater" yaml:"updater" mapstructure:"updater"`
}

func (a *ConfigSvc) SaveState() error {
	// Depending on the extension used, write different format
	switch path.Ext(a.cfgFilePath) {
	case ".json":
		if err := file.StructToJSONFile(a, a.cfgFilePath); err != nil {
			return err
		}
	case ".yaml":
		if err := file.StructToYamlFile(a, a.cfgFilePath); err != nil {
			return err
		}
	}
	return nil
}

// Common handle for parsing app configuration
func (a *ConfigSvc) Process() error {
	if a.cfgFilePath != "" {
		// Use appconfig file from the flag.
		viper.SetConfigFile(a.cfgFilePath)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err == nil {
			// Search appconfig in home directory with name ".viper" (without extension).
			viper.AddConfigPath(home)
			configName := ".viper"
			if a.cfgEnvironmentPrefix != "" {
				configName = "." + a.cfgEnvironmentPrefix
			}
			viper.SetConfigName(configName)
		}
	}

	// Add support for custom unmarshaller
	viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc()))

	viper.AutomaticEnv()
	if a.cfgEnvironmentPrefix != "" {
		viper.SetEnvPrefix(strings.ToUpper(a.cfgEnvironmentPrefix))
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(&a); err != nil {
		return err
	}

	return nil

}
