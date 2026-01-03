package netconfig

import (
	"time"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/relay/dto"
)

type NetSvcConfig struct {
	relay                    dto.RelayInterface
	ExtraHeaders             netdto.ExtraHeaders `json:"extra_headers,omitempty" yaml:"extra_headers,omitempty" mapstructure:"extra_headers"`
	RequestTimeout           time.Duration       `json:"request_timeout,omitempty" yaml:"request_timeout,omitempty" mapstructure:"request_timeout"`
	UserAgent                string              `json:"user_agent,omitempty" yaml:"user_agent,omitempty" mapstructure:"user_agent"`
	BlacklistDomains         []string            `json:"blacklist_domains,omitempty" yaml:"blacklist_domains,omitempty" mapstructure:"blacklist_domains"`
	WhitelistDomains         []string            `json:"whitelist_domains,omitempty" yaml:"whitelist_domains,omitempty" mapstructure:"whitelist_domains"`
	DownloadCallbackInterval time.Duration       `json:"download_callback_interval,omitempty" yaml:"download_callback_interval,omitempty" mapstructure:"download_callback_interval"`
	// PreferCurlDownloads Instead of using imroc/req for downloads, prefer to use curl found on $PATH if available
	PreferCurlDownloads bool `json:"prefer_curl_downloads,omitempty" yaml:"prefer_curl_downloads,omitempty" mapstructure:"prefer_curl_downloads"`
}

func DefaultNetSvcConfig() NetSvcConfig {
	return NetSvcConfig{
		DownloadCallbackInterval: time.Second * 2,
		ExtraHeaders:             make(netdto.ExtraHeaders),
		BlacklistDomains:         make([]string, 0),
		WhitelistDomains:         []string{"github.com"},
	}
}

func (c *NetSvcConfig) AddBlacklistDomain(domain string) *NetSvcConfig {
	c.BlacklistDomains = append(c.BlacklistDomains, domain)
	return c
}
func (c *NetSvcConfig) SetBlacklistDomains(domains []string) *NetSvcConfig {
	c.BlacklistDomains = domains
	return c
}

func (c *NetSvcConfig) AddWhitelistDomain(domain string) *NetSvcConfig {
	c.WhitelistDomains = append(c.WhitelistDomains, domain)
	return c
}

func (c *NetSvcConfig) SetWhitelistDomains(domains []string) *NetSvcConfig {
	c.WhitelistDomains = domains
	return c
}

func (c *NetSvcConfig) AddExtraHeader(key string, value string) *NetSvcConfig {
	c.ExtraHeaders[key] = value
	return c
}

func (c *NetSvcConfig) WithDownloadCallbackDuration(duration time.Duration) *NetSvcConfig {
	c.DownloadCallbackInterval = duration
	return c
}

func (c *NetSvcConfig) WithExtraHeaders(headers netdto.ExtraHeaders) *NetSvcConfig {
	c.ExtraHeaders = headers
	return c
}

func (c *NetSvcConfig) WithPreferCurl(preference bool) *NetSvcConfig {
	c.PreferCurlDownloads = preference
	return c
}

func (c *NetSvcConfig) WithRelay(relay dto.RelayInterface) *NetSvcConfig {
	c.relay = relay
	return c
}
func (c *NetSvcConfig) Relay() dto.RelayInterface {
	return c.relay
}
