package net

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/maps"
	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

var (
	service     *NetSvc
	serviceOnce sync.Once
)

func ProvideNetSvc(cfg *netconfig.NetSvcConfig) *NetSvc {
	serviceOnce.Do(func() {
		service = &NetSvc{
			cfg:            cfg,
			relay:          cfg.Relay(),
			listenersByURL: make(map[string][]chan netdto.TransferNotification),
			transferState:  maps.NewLockableMap[string, netdto.TransferNotification](),
			clients:        make(map[string]netdto.NetClientInterface),
		}
		cfg.Relay().Debug(RlyNetLog{Msg: "Net service started"})
	})
	return service
}
