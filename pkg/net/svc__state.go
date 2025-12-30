package net

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/joy-dx/gophorth/pkg/hydrate"
	"github.com/joy-dx/gophorth/pkg/net/netclient/httpclient"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

func (s *NetSvc) State() *netdto.NetState {

	return &netdto.NetState{
		ExtraHeaders:             s.cfg.ExtraHeaders,
		RequestTimeout:           s.cfg.RequestTimeout,
		UserAgent:                s.cfg.UserAgent,
		BlacklistDomains:         s.cfg.BlacklistDomains,
		WhitelistDomains:         s.cfg.WhitelistDomains,
		DownloadCallbackInterval: s.cfg.DownloadCallbackInterval,
		PreferCurlDownloads:      s.cfg.PreferCurlDownloads,
		TransfersStatus:          s.transferState.GetAll(),
	}
}

func isCurlAvailable() bool {
	_, err := exec.LookPath("curl")
	return err == nil
}

func (s *NetSvc) Hydrate(ctx context.Context) error {
	if hydrateErr := hydrate.NilCheck("net", map[string]interface{}{
		"config":        s.cfg,
		"relay":         s.relay,
		"transferState": s.transferState,
	}); hydrateErr != nil {
		return hydrateErr
	}
	// On Mac, to conform to download security policy, force curl
	if runtime.GOOS == "darwin" {
		s.cfg.WithPreferCurl(true)
	}
	if s.cfg.PreferCurlDownloads && !isCurlAvailable() {
		s.relay.Warn(RlyNetLog{Msg: "Curl set as preference but, not available"})
		s.cfg.WithPreferCurl(false)
	}
	defaultClientCfg := httpclient.DefaultHTTPClientConfig()
	defaultClient := httpclient.NewHTTPClient(netdto.NET_DEFAULT_CLIENT_REF, s.cfg, &defaultClientCfg)
	s.clients[netdto.NET_DEFAULT_CLIENT_REF] = defaultClient

	return nil
}
