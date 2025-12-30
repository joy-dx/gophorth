package logger

import (
	"sync"

	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
)

var (
	service     *LoggerSvc
	serviceOnce sync.Once
)

func ProvideLoggerSvc(cfg *loggerconfig.LoggerConfig) *LoggerSvc {
	serviceOnce.Do(func() {
		service = &LoggerSvc{
			cfg: cfg,
		}
	})
	return service
}
