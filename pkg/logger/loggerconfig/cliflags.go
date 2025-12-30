package loggerconfig

import (
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/spf13/cobra"
)

const ConfigPrefix = "logger"

func CobraAndViper(cmd *cobra.Command) {
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(cmd)
	configBuilder.SetConfigPrefix([]string{ConfigPrefix})
	configBuilder.AddIntParam(options.LoggerKeyPadding, 16, "Right padding characters for debug log key with simple logger")
	configBuilder.AddStringParam(options.LoggerLevel, string(relaydto.Debug), "Limit of console logs to output. debug, info, warn, error")
	configBuilder.AddStringParam(options.LoggerType, "simple", "whether to use simple or structured reporting")
}
