package releaser

import (
	"log/slog"

	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

const RELAY_RELEASE_CHANNEL relaydto.EventChannel = "releaser"

const RELAY_RELEASE_LOG relaydto.EventRef = "release.log"

type RlyReleaserLog struct {
	Msg string `json:"msg"`
}

func (e RlyReleaserLog) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("type", string(e.RelayType())),
	}
}

func (e RlyReleaserLog) Message() string {
	return e.Msg
}

func (e RlyReleaserLog) RelayChannel() relaydto.EventChannel {
	return RELAY_RELEASE_CHANNEL
}

func (e RlyReleaserLog) RelayType() relaydto.EventRef {
	return RELAY_RELEASE_LOG
}
