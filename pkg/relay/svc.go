package relay

import (
	"os"

	"github.com/joy-dx/gophorth/pkg/relay/relayconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
)

// TODO add hooks to the event being emitted
// RelaySvc Wrapper for imroc/req to normalize usage and shorten implementation
type RelaySvc struct {
	sinks []relaydto.RelaySinkInterface
	cfg   *relayconfig.RelaySvcConfig
}

func (r *RelaySvc) RegisterSink(sink relaydto.RelaySinkInterface) {
	r.sinks = append(r.sinks, sink)
}

func (r *RelaySvc) emit(level relaydto.RelayLevel, data relaydto.RelayEventInterface) {

	// dispatch to registered sinks
	for _, sink := range r.sinks {
		switch level {
		case relaydto.Debug:
			sink.Debug(data)
		case relaydto.Info:
			sink.Info(data)
		case relaydto.Warn:
			sink.Warn(data)
		case relaydto.Error:
			sink.Error(data)
		case relaydto.Fatal:
			sink.Fatal(data)
		}
	}
	// After draining all the sinks, exit if fatal
	if level == relaydto.Fatal {
		os.Exit(1)
	}
}

func (r *RelaySvc) Debug(e relaydto.RelayEventInterface) { r.emit(relaydto.Debug, e) }
func (r *RelaySvc) Info(e relaydto.RelayEventInterface)  { r.emit(relaydto.Info, e) }
func (r *RelaySvc) Warn(e relaydto.RelayEventInterface)  { r.emit(relaydto.Warn, e) }
func (r *RelaySvc) Error(e relaydto.RelayEventInterface) { r.emit(relaydto.Error, e) }
func (r *RelaySvc) Fatal(e relaydto.RelayEventInterface) { r.emit(relaydto.Fatal, e) }
