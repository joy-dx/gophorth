package net

import (
	"context"
	"io"
	"time"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

// publishTransferUpdate is the unified notification function
func (s *NetSvc) publishTransferUpdate(state netdto.TransferNotification) {
	s.transferState.Set(state.Destination, state)

	s.muListeners.Lock()
	listeners := s.listenersByURL[state.Source]
	s.muListeners.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- state:
		default: // drop if listener too slow
		}
	}

	if s.relay != nil {
		s.relay.Info(RlyNetDownload{
			Source:      state.Source,
			Destination: state.Destination,
			Status:      state.Status,
			Percentage:  state.Percentage,
			Msg:         state.Message,
		})
	}

	if state.Status == netdto.COMPLETE ||
		state.Status == netdto.ERROR ||
		state.Status == netdto.STOPPED {
		s.TransferListenerClose(state.Source)
	}
}

type progressReader struct {
	ctx        context.Context
	reader     io.Reader
	total      int64
	readSoFar  int64
	lastReport time.Time
	lastBytes  int64
	interval   time.Duration
	startTime  time.Time
	onProgress func(downloaded, total int64, percent float64, speed float64, eta time.Duration)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	select {
	case <-pr.ctx.Done():
		return 0, pr.ctx.Err()
	default:
	}

	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.readSoFar += int64(n)
		now := time.Now()
		if now.Sub(pr.lastReport) >= pr.interval {
			deltaBytes := pr.readSoFar - pr.lastBytes
			deltaTime := now.Sub(pr.lastReport).Seconds()
			speed := float64(deltaBytes) / deltaTime // bytes/sec

			var pct float64
			if pr.total > 0 {
				pct = float64(pr.readSoFar) / float64(pr.total) * 100
				if pct > 100 {
					pct = 100
				}
			}

			var eta time.Duration
			if pr.total > 0 && speed > 0 {
				remaining := float64(pr.total - pr.readSoFar)
				eta = time.Duration(remaining/speed) * time.Second
			}

			pr.onProgress(pr.readSoFar, pr.total, pct, speed, eta)
			pr.lastReport = now
			pr.lastBytes = pr.readSoFar
		}
	}

	return n, err
}
