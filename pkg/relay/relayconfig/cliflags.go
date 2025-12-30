package relayconfig

import (
	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/spf13/cobra"
)

const ConfigPrefix = "relay"

func CobraAndViper(cmd *cobra.Command) {
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(cmd)
	configBuilder.SetConfigPrefix([]string{ConfigPrefix})
}
