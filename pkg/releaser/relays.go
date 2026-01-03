package releaser

import (
	"log/slog"

	"github.com/joy-dx/relay/dto"
)

const RELAY_RELEASE_CHANNEL dto.EventChannel = "releaser"

const RELAY_RELEASE_LOG dto.EventRef = "release.log"

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

func (e RlyReleaserLog) RelayChannel() dto.EventChannel {
	return RELAY_RELEASE_CHANNEL
}

func (e RlyReleaserLog) RelayType() dto.EventRef {
	return RELAY_RELEASE_LOG
}
