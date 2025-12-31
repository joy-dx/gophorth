package guisinks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const WailsSinkRef = "wails"

var levels = []relaydto.RelayLevel{relaydto.Fatal, relaydto.Error, relaydto.Warn, relaydto.Info, relaydto.Debug}

// ValueIndexFromSlice Simple checker to see if an entry exists within a list
func ValueIndexFromSlice[T comparable](a T, list []T) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return -1
}

type WailsSink struct {
	level int
	cfg   *relaydto.RelaySinkConfig
	ctx   context.Context
}

func NewWailsSink(ctx context.Context, cfg *relaydto.RelaySinkConfig) *WailsSink {
	levelIndex := ValueIndexFromSlice(cfg.Level, levels)
	return &WailsSink{
		ctx:   ctx,
		cfg:   cfg,
		level: levelIndex,
	}
}

func (s *WailsSink) Ref() string {
	return WailsSinkRef
}

func (s *WailsSink) emit(e relaydto.RelayEventInterface, level relaydto.RelayLevel) {
	marshaledData, _ := json.Marshal(e)
	event := relaydto.Event{
		Channel:   e.RelayChannel(),
		Ref:       e.RelayType(),
		Level:     level,
		Timestamp: time.Now(),
		Data:      marshaledData,
	}
	runtime.EventsEmit(s.ctx, string(e.RelayChannel()), event)
}

func (s *WailsSink) Debug(e relaydto.RelayEventInterface) {
	if s.level > 3 {
		s.emit(e, relaydto.Debug)
	}
}
func (s *WailsSink) Info(e relaydto.RelayEventInterface) {
	if s.level > 2 {
		s.emit(e, relaydto.Info)
	}
}
func (s *WailsSink) Warn(e relaydto.RelayEventInterface) {
	if s.level > 1 {
		s.emit(e, relaydto.Warn)
	}
}
func (s *WailsSink) Error(e relaydto.RelayEventInterface) {
	s.emit(e, relaydto.Error)
}
func (s *WailsSink) Fatal(e relaydto.RelayEventInterface) {
	s.emit(e, relaydto.Fatal)
}

func (s *WailsSink) Meta(e relaydto.RelayEventInterface) {

}
