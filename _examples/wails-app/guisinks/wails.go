package guisinks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/sinks"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const WailsSinkRef = "wails"

type WailsSink struct {
	level int
	cfg   *dto.RelaySinkConfig
	ctx   context.Context
}

func NewWailsSink(ctx context.Context, cfg *dto.RelaySinkConfig) *WailsSink {
	levelIndex := sinks.GetLogLevelIndex(cfg.Level, dto.Levels)
	return &WailsSink{
		ctx:   ctx,
		cfg:   cfg,
		level: levelIndex,
	}
}

func (s *WailsSink) Ref() string {
	return WailsSinkRef
}

func (s *WailsSink) emit(e dto.RelayEventInterface, level dto.RelayLevel) {
	marshaledData, _ := json.Marshal(e)
	event := dto.Event{
		Channel:   e.RelayChannel(),
		Ref:       e.RelayType(),
		Level:     level,
		Timestamp: time.Now(),
		Data:      marshaledData,
	}
	runtime.EventsEmit(s.ctx, string(e.RelayChannel()), event)
}

func (s *WailsSink) Debug(e dto.RelayEventInterface) {
	if s.level <= 3 {
		return
	}
	s.emit(e, dto.Debug)
}

func (s *WailsSink) Info(e dto.RelayEventInterface) {
	if s.level <= 2 {
		return
	}
	s.emit(e, dto.Info)
}

func (s *WailsSink) Warn(e dto.RelayEventInterface) {
	if s.level <= 1 {
		return
	}
	s.emit(e, dto.Warn)
}

func (s *WailsSink) Error(e dto.RelayEventInterface) {
	s.emit(e, dto.Error)
}

func (s *WailsSink) Fatal(e dto.RelayEventInterface) {
	s.emit(e, dto.Fatal)
}

func (s *WailsSink) Meta(e dto.RelayEventInterface) {

}
