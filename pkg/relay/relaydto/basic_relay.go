package relaydto

import (
	"log/slog"
)

const RELAY_CHANNEL EventChannel = "relay"

const RELAY_LOG EventRef = "relay.log"

type RlyLog struct {
	Msg string `json:"msg" yaml:"msg"`
}

func (e RlyLog) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("msg", e.Msg),
	}
}

func (e RlyLog) Message() string {
	return e.Msg
}

func (e RlyLog) RelayChannel() EventChannel {
	return RELAY_CHANNEL
}

func (e RlyLog) RelayType() EventRef {
	return RELAY_LOG
}
