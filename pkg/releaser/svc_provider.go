package releaser

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

var (
	service     *ReleaserSvc
	serviceOnce sync.Once
)

func ProvideReleaserSvc(cfg *releaserdto.ReleaserConfig) *ReleaserSvc {
	serviceOnce.Do(func() {
		service = &ReleaserSvc{
			cfg:   cfg,
			relay: cfg.Relay,
		}
	})
	return service
}
