package net

import (
	"log/slog"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/relay/dto"
)

const RELAY_NET_CHANNEL dto.EventChannel = "net"

const RELAY_NET_DOWNLOAD dto.EventRef = "net.download"

type RlyNetDownload struct {
	Source      string                `json:"source" yaml:"source"`
	Destination string                `json:"destination" yaml:"destination"`
	Msg         string                `json:"msg,omitempty" yaml:"msg,omitempty"`
	Status      netdto.TransferStatus `json:"status" yaml:"status"`
	Percentage  float64               `json:"percentage" yaml:"percentage"`
	// TotalSize length content in bytes. The value -1 indicates that the length is unknown
	TotalSize int64 `json:"total_size,omitempty" yaml:"total_size,omitempty"`
	// Downloaded downloaded body length in bytes
	Downloaded int64 `json:"downloaded,omitempty" yaml:"downloaded,omitempty"`
}

func (e RlyNetDownload) ToSlog() []slog.Attr {
	zapFields := []slog.Attr{
		slog.String("type", string(e.RelayType())),
		slog.String("src", e.Source),
		slog.String("dst", e.Destination),
		slog.String("status", string(e.Status)),
		slog.Float64("percentage", e.Percentage),
	}
	if e.Downloaded > 0 {
		zapFields = append(zapFields, slog.Int64("downloaded", e.Downloaded))
	}
	if e.TotalSize > 0 {
		zapFields = append(zapFields, slog.Int64("total_size", e.TotalSize))
	}
	return zapFields
}

func (e RlyNetDownload) Message() string {
	return e.Msg
}

func (e RlyNetDownload) RelayChannel() dto.EventChannel {
	return RELAY_NET_CHANNEL
}

func (e RlyNetDownload) RelayType() dto.EventRef {
	return RELAY_NET_DOWNLOAD
}

const RELAY_NET_LOG dto.EventRef = "net.log"

type RlyNetLog struct {
	Msg string `json:"msg"`
}

func (e RlyNetLog) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("type", string(e.RelayType())),
	}
}

func (e RlyNetLog) Message() string {
	return e.Msg
}

func (e RlyNetLog) RelayChannel() dto.EventChannel {
	return RELAY_NET_CHANNEL
}

func (e RlyNetLog) RelayType() dto.EventRef {
	return RELAY_NET_LOG
}
