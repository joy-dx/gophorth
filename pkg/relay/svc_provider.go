package relay

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/relay/relayconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

var (
	service     *RelaySvc
	serviceOnce sync.Once
)

func ProvideRelaySvc(cfg *relayconfig.RelaySvcConfig) *RelaySvc {
	serviceOnce.Do(func() {
		service = &RelaySvc{
			cfg:   cfg,
			sinks: make([]relaydto.RelaySinkInterface, 0),
		}
	})
	return service
}
