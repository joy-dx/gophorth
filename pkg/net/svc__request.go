package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/joy-dx/gophorth/pkg/delay"
	"github.com/joy-dx/gophorth/pkg/net/netclient/httpclient"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/net/netutils"
)

// Get RequestWithRetry
func (s *NetSvc) Get(ctx context.Context, url string, withRetry bool) (netdto.Response, error) {
	httpRequestConfig := httpclient.DefaultHTTPRequestConfig()
	httpRequestConfig.WithURL(url)
	cfg := netdto.DefaultRequestConfig()
	cfg.WithReqConfig(&httpRequestConfig).
		WithTaskName("GET " + url)

	if withRetry {
		return s.RequestWithRetry(ctx, &cfg)
	}
	return s.RequestOnce(ctx, &cfg)

}

// Post RequestWithRetry
func (s *NetSvc) Post(ctx context.Context, url string, payload map[string]interface{}, withRetry bool) (netdto.Response, error) {
	httpRequestConfig := httpclient.DefaultHTTPRequestConfig()
	httpRequestConfig.WithURL(url).
		WithBody(payload).
		WithMethod(http.MethodPost)
	cfg := netdto.DefaultRequestConfig()
	cfg.WithReqConfig(&httpRequestConfig).
		WithTaskName("POST " + url)

	if withRetry {
		return s.RequestWithRetry(ctx, &cfg)
	}
	return s.RequestOnce(ctx, &cfg)
}

func (s *NetSvc) RequestWithRetry(ctx context.Context, cfg *netdto.RequestConfig) (netdto.Response, error) {
	if cfg == nil {
		return netdto.Response{}, errors.New("nil RequestConfig provided")
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.Delay == nil {
		cfg.Delay = delay.ConstantDelay{Period: 1}
	}
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			cfg.Delay.Wait(cfg.TaskName, attempt)
		}

		resp, err := s.RequestOnce(ctx, cfg)
		if err != nil {
			lastErr = err
			// transient network errors → retry
			if netutils.IsTemporaryErr(err) && attempt < cfg.MaxRetries {
				continue
			}
			return resp, err
		}

		// Retry for 5xx responses only if retries remain
		if resp.StatusCode >= 500 && attempt < cfg.MaxRetries {
			lastErr = fmt.Errorf("server error (%d)", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return netdto.Response{}, fmt.Errorf("failed after %d attempts: %w", cfg.MaxRetries+1, lastErr)
}

func (s *NetSvc) RequestOnce(ctx context.Context, cfg *netdto.RequestConfig) (netdto.Response, error) {

	if cfg.ClientRef == "" {
		return netdto.Response{}, errors.New("nil ClientRef provided")
	}

	if cfg.TaskName == "" {
		cfg.TaskName = "http_request"
	}

	netClient, isOK := s.clients[cfg.ClientRef]
	if !isOK {
		return netdto.Response{}, fmt.Errorf("client not found: %s", cfg.ClientRef)
	}

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	response, err := netClient.ProcessRequest(ctx, cfg)
	if err != nil {
		return netdto.Response{}, fmt.Errorf("perform request: %w", err)
	}

	if cfg.ResponseObject != nil && len(response.Body) > 0 {
		if unmarshalErr := json.Unmarshal(response.Body, cfg.ResponseObject); unmarshalErr != nil {
			return response, fmt.Errorf("unmarshal response: %w", unmarshalErr)
		}
	}

	return response, nil
}
