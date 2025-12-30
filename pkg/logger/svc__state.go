package logger

import (
	"fmt"

	"github.com/joy-dx/gophorth/pkg/hydrate"
	"github.com/joy-dx/gophorth/pkg/logger/loggersinks"
)

func (s *LoggerSvc) Hydrate() error {
	if hydrateErr := hydrate.NilCheck("logger", map[string]interface{}{
		"config": s.cfg,
	}); hydrateErr != nil {
		return hydrateErr
	}

	switch s.cfg.Type {
	case loggersinks.SimpleLoggerRef:
		s.logger = loggersinks.NewSimpleLogger(s.cfg)
	case loggersinks.StructuredLoggerRef:
		s.logger = loggersinks.NewStructuredLogger(s.cfg)
	default:
		return fmt.Errorf("unknown logger type: %s", s.cfg.Type)
	}
	return nil
}
