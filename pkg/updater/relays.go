package updater

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

const RELAY_UPDATER_CHANNEL relaydto.EventChannel = "updater"

const RELAY_UPDATER_LOG relaydto.EventRef = "updater.log"

type RlyUpdaterLog struct {
	Msg string `json:"msg"`
}

func (e RlyUpdaterLog) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("type", string(e.RelayType())),
	}
}

func (e RlyUpdaterLog) Message() string {
	return e.Msg
}

func (e RlyUpdaterLog) RelayChannel() relaydto.EventChannel {
	return RELAY_UPDATER_CHANNEL
}

func (e RlyUpdaterLog) RelayType() relaydto.EventRef {
	return RELAY_UPDATER_LOG
}

const RELAY_UPDATER_NEW_VERSION relaydto.EventRef = "updater.new_version"

type RlyNewVersion struct {
	ReleasedAt *time.Time `json:"released_at"`
	Version    string     `json:"version"`
}

func (e RlyNewVersion) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("released_at", e.ReleasedAt.Format(time.RFC3339)),
		slog.String("version", e.Version),
	}
}

func (e RlyNewVersion) Message() string {
	return fmt.Sprintf("app version %s available, released on %s", e.Version, e.ReleasedAt.Format(time.RFC3339))
}

func (e RlyNewVersion) RelayChannel() relaydto.EventChannel {
	return RELAY_UPDATER_CHANNEL
}

func (e RlyNewVersion) RelayType() relaydto.EventRef {
	return RELAY_UPDATER_NEW_VERSION
}
