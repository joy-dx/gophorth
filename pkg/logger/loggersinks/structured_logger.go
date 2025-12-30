package loggersinks

import (
	"context"
	"log/slog"
	"os"

	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

const StructuredLoggerRef = "structured"

type StructuredLogger struct {
	logger *slog.Logger
	cfg    *loggerconfig.LoggerConfig
}

func (s *StructuredLogger) Ref() string {
	return StructuredLoggerRef
}

func NewStructuredLogger(cfg *loggerconfig.LoggerConfig) *StructuredLogger {

	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: convertLevel(cfg.Level),
	})
	return &StructuredLogger{
		cfg:    cfg,
		logger: slog.New(h),
	}
}

func convertLevel(l relaydto.RelayLevel) slog.Level {
	switch l {
	case relaydto.Debug:
		return slog.LevelDebug
	case relaydto.Info:
		return slog.LevelInfo
	case relaydto.Warn:
		return slog.LevelWarn
	case relaydto.Error:
		return slog.LevelError
	case relaydto.Fatal:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func (s *StructuredLogger) Debug(e relaydto.RelayEventInterface) {
	s.logger.LogAttrs(context.Background(), slog.LevelDebug, e.Message(), e.ToSlog()...)
}
func (s *StructuredLogger) Info(e relaydto.RelayEventInterface) {
	s.logger.LogAttrs(context.Background(), slog.LevelInfo, e.Message(), e.ToSlog()...)
}
func (s *StructuredLogger) Warn(e relaydto.RelayEventInterface) {
	s.logger.LogAttrs(context.Background(), slog.LevelWarn, e.Message(), e.ToSlog()...)
}
func (s *StructuredLogger) Error(e relaydto.RelayEventInterface) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, e.Message(), e.ToSlog()...)
}
func (s *StructuredLogger) Fatal(e relaydto.RelayEventInterface) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, "FATAL", e.ToSlog()...)
}
