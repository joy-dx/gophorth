package loggerdto

import "github.com/joy-dx/gophorth/pkg/relay/relaydto"

type LoggerSvc interface {
	GetLogger() relaydto.RelaySinkInterface
	Hydrate() error
}
