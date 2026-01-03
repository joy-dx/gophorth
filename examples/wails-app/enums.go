package main

import (
	"github.com/joy-dx/gophorth/pkg/net"
	"github.com/joy-dx/gophorth/pkg/releaser"
	"github.com/joy-dx/gophorth/pkg/updater"
	"github.com/joy-dx/relay/dto"
	"github.com/joy-dx/relay/events"
)

type Channel string

// Channels For exporting to the frontend
var Channels = []struct {
	Value  Channel
	TSName string
}{
	{Channel(net.RELAY_NET_CHANNEL), "RELAY_NET"},
	{Channel(events.RELAY_CHANNEL), "RELAY_BASE"},
	{Channel(releaser.RELAY_RELEASE_CHANNEL), "RELAY_RELEASER"},
	{Channel(updater.RELAY_UPDATER_CHANNEL), "RELAY_UPDATER"},
}

type Relay dto.EventRef

var Relays = []struct {
	Value  Relay
	TSName string
}{
	{Relay(net.RELAY_NET_DOWNLOAD), "NET_DOWNLOAD"},
	{Relay(net.RELAY_NET_LOG), "NET_LOG"},
	{Relay(events.RELAY_LOG), "RELAY_LOG"},
	{Relay(releaser.RELAY_RELEASE_LOG), "RELEASE_LOG"},
	{Relay(updater.RELAY_UPDATER_LOG), "UPDATER_LOG"},
	{Relay(updater.RELAY_UPDATER_NEW_VERSION), "UPDATER_NEW_VERSION"},
}
