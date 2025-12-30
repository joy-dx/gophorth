package netconfig

import (
	"time"

	"github.com/joy-dx/gophorth/pkg/config/builder"
	"github.com/joy-dx/gophorth/pkg/config/options"
	"github.com/spf13/cobra"
)

const ConfigPrefix = "net"

func CobraAndViper(cmd *cobra.Command) {
	configBuilder := builder.ConfigBuilder{}
	configBuilder.SetCommand(cmd)
	configBuilder.SetConfigPrefix([]string{ConfigPrefix})
	configBuilder.AddStringMapParam(options.NetExtraHeaders, map[string]string{}, "Extra headers to include in requests e.g. {\"x-agent\": \"joydx\"")
	configBuilder.AddDurationParam(options.NetDownloadCallbackInterval, time.Duration(2*time.Second), "Download polling interval to report on download progress")
	configBuilder.AddDurationParam(options.NetRequestTimeout, time.Duration(300*time.Second), "Time allowed in seconds for net requests to take before timing out")
	configBuilder.AddStringParam(options.NetUserAgent, "joydx", "Default user agent to include in header when making net requests")
	configBuilder.AddStringSliceParam(options.NetDomainBlacklist, []string{}, "Blacklisted domains that will error when trying to make net requests to them")
	configBuilder.AddStringSliceParam(options.NetDomainWhitelist, []string{"github.com"}, "Domain whitelist for net requests")
}
