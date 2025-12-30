package net

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	buffer "github.com/joy-dx/gophorth/pkg/buffers"
	"github.com/joy-dx/gophorth/pkg/cryptography"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/stringz"
)

// downloadFileWithHTTP streams via net/http with progress
func (s *NetSvc) downloadFileWithHTTP(
	ctx context.Context,
	cfg *netdto.DownloadFileConfig,
	destination string,
) error {
	s.relay.Debug(RlyNetDownload{
		Source:      cfg.URL,
		Destination: destination,
		Msg:         "Downloading via net/http",
		Status:      netdto.IN_PROGRESS,
	})

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("could not create destination folder %q: %w", destination, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("bad HTTP status: %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("could not create output file %q: %w", destination, err)
	}
	defer out.Close()

	total := resp.ContentLength
	if total <= 0 {
		s.relay.Warn(RlyNetDownload{Source: cfg.URL, Msg: "unknown file size"})
	}

	interval := s.cfg.DownloadCallbackInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}

	report := func(downloaded, total int64, percent float64, speed float64, eta time.Duration) {
		s.publishTransferUpdate(netdto.TransferNotification{
			Source:      cfg.URL,
			Destination: destination,
			Status:      netdto.IN_PROGRESS,
			Downloaded:  downloaded,
			TotalSize:   total,
			Percentage:  percent,
		})
	}

	pr := &progressReader{
		ctx:        ctx,
		reader:     resp.Body,
		total:      total,
		interval:   interval,
		lastReport: time.Now(),
		startTime:  time.Now(),
		onProgress: report,
	}

	buf := make([]byte, 64*1024)
	_, err = io.CopyBuffer(out, pr, buf)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.publishTransferUpdate(netdto.TransferNotification{
				Source:      cfg.URL,
				Destination: destination,
				Status:      netdto.STOPPED,
			})
			return ctx.Err()
		}

		s.publishTransferUpdate(netdto.TransferNotification{
			Source:      cfg.URL,
			Destination: destination,
			Status:      netdto.ERROR,
			Message:     err.Error(),
		})
		return fmt.Errorf("file transfer failed for %s: %w", cfg.URL, err)
	}

	if cfg.Checksum != "" {
		checkErr := cryptography.Sha256SumVerify(destination, cfg.Checksum)
		if checkErr != nil {
			s.publishTransferUpdate(netdto.TransferNotification{
				Source:      cfg.URL,
				Destination: destination,
				Status:      netdto.ERROR,
				Percentage:  100,
				Message:     "failed to verify checksum",
			})
			return fmt.Errorf("checksum verification failed: %w", checkErr)
		}
	}

	s.publishTransferUpdate(netdto.TransferNotification{
		Source:      cfg.URL,
		Destination: destination,
		Status:      netdto.COMPLETE,
		Downloaded:  total,
		TotalSize:   total,
		Percentage:  100,
		Message:     "download complete",
	})
	return nil
}

// =====================================================================
// Curl Downloader Implementation
// =====================================================================

func (s *NetSvc) downloadFileWithCurl(
	ctx context.Context,
	cfg *netdto.DownloadFileConfig,
	destination string,
) error {
	s.relay.Debug(RlyNetDownload{
		Source:      cfg.URL,
		Destination: destination,
		Msg:         "Downloading via curl",
		Status:      netdto.IN_PROGRESS,
	})

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("could not create destination folder %q: %w", destination, err)
	}

	curlCmd := exec.CommandContext(ctx, "curl", "-L", "--progress-bar", "-o", destination, cfg.URL)
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	curlCmd.Stdout = stdoutBuf
	curlCmd.Stderr = stderrBuf

	if err := curlCmd.Start(); err != nil {
		return fmt.Errorf("failed to start curl: %w", err)
	}

	interval := s.cfg.DownloadCallbackInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	done := make(chan error, 1)

	go func() {
		err := curlCmd.Wait()
		select {
		case done <- err:
		case <-ctx.Done():
		}
	}()

	for {
		select {
		case <-ticker.C:
			if msg, newContent := buffer.Flush(stderrBuf); newContent && len(msg) >= 6 {
				if parsed, err := stringz.ParsePercentage(msg[len(msg)-6:]); err == nil {
					if parsed > 100 {
						parsed = 100
					}
					s.publishTransferUpdate(netdto.TransferNotification{
						Source:      cfg.URL,
						Destination: destination,
						Status:      netdto.IN_PROGRESS,
						Percentage:  parsed,
					})
				}
			}
		case <-ctx.Done():
			if curlCmd.Process != nil {
				_ = curlCmd.Process.Kill()
			}
			s.publishTransferUpdate(netdto.TransferNotification{
				Source:      cfg.URL,
				Destination: destination,
				Status:      netdto.STOPPED,
			})
			return ctx.Err()

		case err := <-done:
			ticker.Stop()
			if err != nil {
				s.publishTransferUpdate(netdto.TransferNotification{
					Source:      cfg.URL,
					Destination: destination,
					Status:      netdto.ERROR,
					Message:     err.Error(),
				})
				return fmt.Errorf("curl download failed: %w", err)
			}

			if cfg.Checksum != "" {
				checkErr := cryptography.Sha256SumVerify(destination, cfg.Checksum)
				if checkErr != nil {
					s.publishTransferUpdate(netdto.TransferNotification{
						Source:      cfg.URL,
						Destination: destination,
						Status:      netdto.ERROR,
						Percentage:  100,
						Message:     "failed to verify checksum",
					})
					return fmt.Errorf("checksum verification failed: %w", checkErr)
				}
			}

			s.publishTransferUpdate(netdto.TransferNotification{
				Source:      cfg.URL,
				Destination: destination,
				Status:      netdto.COMPLETE,
				Percentage:  100,
				Message:     "download complete",
			})
			return nil
		}
	}
}

func (s *NetSvc) DownloadFile(ctx context.Context, cfg *netdto.DownloadFileConfig) (string, error) {

	if cfg.OutputFileName == "" {
		// Try and get the filename from the URL and use the destination folder instead
		filename, err := stringz.FilenameFromUrl(cfg.URL)
		if err != nil {
			return "", err
		}
		cfg.OutputFileName = filename
	}

	destination := filepath.Join(cfg.DestinationFolder, cfg.OutputFileName)

	s.relay.Info(RlyNetDownload{
		Source:      cfg.URL,
		Destination: destination,
		Status:      netdto.IN_PROGRESS,
		Percentage:  0,
		Msg:         fmt.Sprintf("starting download: %s", cfg.URL),
	})

	if s.cfg.PreferCurlDownloads {
		return destination, s.downloadFileWithCurl(ctx, cfg, destination)
	}

	return destination, s.downloadFileWithHTTP(ctx, cfg, destination)
}
