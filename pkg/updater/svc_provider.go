package updater

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
)

var (
	service     *UpdaterSvc
	serviceOnce sync.Once
)

func ProvideUpdaterSvc(cfg *updaterdto.UpdaterConfig) *UpdaterSvc {
	serviceOnce.Do(func() {
		service = &UpdaterSvc{
			cfg:    cfg,
			netSvc: cfg.NetSvc,
			relay:  cfg.Relay,
			status: updaterdto.INITIAL,
		}
	})
	return service
}
