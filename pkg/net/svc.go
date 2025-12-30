package net

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/maps"
	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

// NetSvc Wrapper for imroc/req to normalize usage and shorten implementation
type NetSvc struct {
	cfg            *netconfig.NetSvcConfig
	relay          relaydto.RelayInterface
	clients        map[string]netdto.NetClientInterface
	transferState  maps.Lockable[string, netdto.TransferNotification]
	muListeners    sync.Mutex
	listenersByURL map[string][]chan netdto.TransferNotification
}

func (s *NetSvc) RegisterClient(ref string, client netdto.NetClientInterface) {
	s.clients[ref] = client
}

// TransferListener returns a channel of updates for a particular URL
func (s *NetSvc) TransferListener(sourceURL string) <-chan netdto.TransferNotification {
	s.muListeners.Lock()
	defer s.muListeners.Unlock()

	ch := make(chan netdto.TransferNotification, 10)
	s.listenersByURL[sourceURL] = append(s.listenersByURL[sourceURL], ch)
	return ch
}

// TransferListenerClose closes all channels for a given URL manually
func (s *NetSvc) TransferListenerClose(sourceURL string) {
	s.muListeners.Lock()
	defer s.muListeners.Unlock()
	if chans, ok := s.listenersByURL[sourceURL]; ok {
		for _, c := range chans {
			close(c)
		}
		delete(s.listenersByURL, sourceURL)
	}
}
