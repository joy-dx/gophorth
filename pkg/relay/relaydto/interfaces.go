package relaydto

import (
	"log/slog"
)

type RelayEventInterface interface {
	RelayChannel() EventChannel
	RelayType() EventRef
	Message() string
	ToSlog() []slog.Attr
}

type RelayInterface interface {
	Debug(data RelayEventInterface)
	Info(data RelayEventInterface)
	Warn(data RelayEventInterface)
	Error(data RelayEventInterface)
	Fatal(data RelayEventInterface)
}

type RelaySinkInterface interface {
	Ref() string
	Debug(data RelayEventInterface)
	Info(data RelayEventInterface)
	Warn(data RelayEventInterface)
	Error(data RelayEventInterface)
	Fatal(data RelayEventInterface)
}
