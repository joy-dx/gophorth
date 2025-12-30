package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/joy-dx/gophorth/pkg/net/netconfig"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/net/netutils"
)

// -----------------------------------------------------------------------------
// PERSISTENT CLIENT IMPLEMENTATION
// -----------------------------------------------------------------------------

// HTTPClient is a high-level wrapper around netdto.NetInterface,
// providing automatic authentication and session management.
//
// It supports multiple authentication modes:
//   - OAuth2 TokenSource (golang.org/x/oauth2)
//   - Custom AuthProvider
//   - Cookie-based sessions
//
// HTTPClient is suitable for long-lived service integrations where
// multiple requests share authentication state safely.

const NetClientHTTPRef netdto.NetClientType = "net.client.http"

type HTTPClient struct {
	NetClient netdto.NetClient `json:"net_client" yaml:"net_client"`
	cfg       *HTTPClientConfig
	netCfg    *netconfig.NetSvcConfig
	client    *http.Client
	token     netdto.TokenInfo
	tokenMu   sync.RWMutex
}

func NewHTTPClient(ref string, netCfg *netconfig.NetSvcConfig, cfg *HTTPClientConfig) *HTTPClient {
	return &HTTPClient{
		cfg: cfg,
		NetClient: netdto.NetClient{
			Name:        "HTTP Client",
			Ref:         ref,
			ClientType:  NetClientHTTPRef,
			Description: "Perform HTTP requests to given URLs including auth support",
		},
		client: &http.Client{
			Timeout: netCfg.RequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

func (c *HTTPClient) Ref() string {
	return c.NetClient.Ref
}
func (c *HTTPClient) Type() netdto.NetClientType {
	return NetClientHTTPRef
}

// -----------------------------------------------------------------------------
// REQUEST EXECUTION
// -----------------------------------------------------------------------------
// RequestWithRetry executes one authenticated, middleware-wrapped call through the underlying netdto.NetInterface.
// Automatically handles token Lifetimes, OAuth2 renewal, and cookie sessions.
//
// If multiple authentication mechanisms are configured, OAuth2 takes precedence.
// AuthProvider is used as a fallback.
func (c *HTTPClient) ProcessRequest(ctx context.Context, inCfg *netdto.RequestConfig) (netdto.Response, error) {
	cfg, castOk := inCfg.ReqConfig.(*HTTPRequestConfig)
	if !castOk {
		return netdto.Response{}, errors.New("problem casting to httprequestconfig")
	}

	for _, mw := range c.cfg.Middlewares {
		if err := mw(ctx, inCfg); err != nil {
			return netdto.Response{}, fmt.Errorf("middleware aborted: %w", err)
		}
	}

	if err := c.ensureToken(ctx); err != nil {
		return netdto.Response{}, fmt.Errorf("ensure token: %w", err)
	}

	// Step 3: attach credentials (Authorization or Cookies)
	c.tokenMu.RLock()
	c.attachAuth(cfg)
	c.tokenMu.RUnlock()

	// Prepare request body once as reusable []byte to safely rebuild request
	bodyBuf, contentType, err := netutils.PrepareBody(cfg.Body, cfg.BodyType)
	if err != nil {
		return netdto.Response{}, fmt.Errorf("prepare body: %w", err)
	}

	// BuildID a new request using bytes.NewReader to ensure we can safely retry internally if needed
	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.URL, bytes.NewReader(bodyBuf))
	if err != nil {
		return netdto.Response{}, fmt.Errorf("create request: %w", err)
	}

	// Apply headers safely — don’t override user defined Authorization
	for k, v := range cfg.Headers {
		if k == "Authorization" && req.Header.Get("Authorization") != "" {
			continue
		}
		req.Header.Set(k, v)
	}

	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Defensive client.Do handling — httpResp may be non-nil with error
	httpResp, reqErr := c.client.Do(req)
	if httpResp != nil {
		defer func() {
			io.Copy(io.Discard, httpResp.Body) // drain fully for connection reuse
			httpResp.Body.Close()
		}()
	}
	if reqErr != nil {
		return netdto.Response{}, fmt.Errorf("perform request: %w", reqErr)
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return netdto.Response{}, fmt.Errorf("read body: %w", err)
	}

	response := netdto.Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header.Clone(),
		Body:       bodyBytes,
	}

	// Capture cookies, prunes if expired
	if setCookies := response.Headers["Set-Cookie"]; len(setCookies) > 0 {
		c.captureCookiesFromResponse(response)
	}

	// Guard unauthorized error type explicitly
	if response.StatusCode == http.StatusUnauthorized {
		return response, fmt.Errorf("unauthorized: %s", cfg.URL)
	}

	return response, nil
}
