package logger

import (
	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

// LoggerSvc A factory for generating the app logger relay sink
type LoggerSvc struct {
	cfg    *loggerconfig.LoggerConfig
	logger relaydto.RelaySinkInterface
}

func (s *LoggerSvc) GetLogger() relaydto.RelaySinkInterface {
	return s.logger
}
