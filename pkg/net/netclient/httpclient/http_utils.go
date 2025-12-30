package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

// ensureToken verifies if an active token is valid, auto-refreshing if necessary.
func (c *HTTPClient) ensureToken(ctx context.Context) error {
	c.tokenMu.RLock()
	valid := !c.token.IsExpired(c.cfg.RefreshBuffer)
	c.tokenMu.RUnlock()
	if valid {
		return nil
	}
	return c.refreshToken(ctx)
}

// refreshToken retrieves a new token using OAuth2 or AuthProvider.
func (c *HTTPClient) refreshToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if !c.token.IsExpired(c.cfg.RefreshBuffer) {
		return nil
	}

	// Case 1: OAuth2 integration
	if c.cfg.OAuthSource != nil {
		oauthTok, err := c.cfg.OAuthSource.Token()
		if err != nil {
			return fmt.Errorf("oauth2 token fetch: %w", err)
		}
		c.token.AccessToken = oauthTok.AccessToken
		c.token.TokenType = normalizeAuthType(oauthTok.TokenType)
		c.token.Expiry = oauthTok.Expiry
		return nil
	}

	// Case 2: custom AuthProvider
	if c.cfg.AuthProvider != nil {
		var newTok netdto.TokenInfo
		var res *http.Response
		var err error

		if c.token.AccessToken == "" && len(c.token.Cookies) == 0 {
			newTok, err = c.cfg.AuthProvider.Authenticate(ctx)
		} else {
			newTok, err = c.cfg.AuthProvider.Refresh(ctx, c.token)
			if err != nil {
				newTok, err = c.cfg.AuthProvider.Authenticate(ctx)
			}
		}
		if err != nil {
			return fmt.Errorf("auth provider refresh: %w", err)
		}

		// Infer token type
		newTok.TokenType = normalizeAuthType(newTok.TokenType)

		// Capture cookies if any
		if res != nil && len(res.Header["Set-Cookie"]) > 0 {
			newTok.Cookies = res.Cookies()
		}
		c.token = newTok
		return nil
	}

	// Case 3: no auth
	return nil
}

// normalizeAuthType ensures proper "Bearer", "Basic", or custom capitalization.
func normalizeAuthType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "bearer":
		return "Bearer"
	case "basic":
		return "Basic"
	default:
		if t == "" {
			return "Bearer"
		}
		return t
	}
}

// -----------------------------------------------------------------------------
// HEADER + COOKIE MANAGEMENT
// -----------------------------------------------------------------------------

// attachAuth injects auth credentials or cookies into a netdto.RequestConfig.
func (c *HTTPClient) attachAuth(cfg *HTTPRequestConfig) {
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}

	if c.token.AccessToken != "" {
		authHeader := fmt.Sprintf("%s %s", normalizeAuthType(c.token.TokenType), c.token.AccessToken)
		cfg.Headers["Authorization"] = authHeader
		return
	}

	if len(c.token.Cookies) > 0 {
		merged := ""
		for _, c := range c.token.Cookies {
			merged += c.String() + "; "
		}
		cfg.Headers["Cookie"] = merged
	}
}

// captureCookiesFromnetdto.Response stores updated cookies from Set-Cookie headers.
func (c *HTTPClient) captureCookiesFromResponse(resp netdto.Response) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	for _, set := range resp.Headers["Set-Cookie"] {
		cookies := parseSetCookieHeader(set)
		for _, cookie := range cookies {
			c.storeOrReplaceCookie(cookie)
		}
	}
}

// storeOrReplaceCookie updates or appends a cookie by its name.
func (c *HTTPClient) storeOrReplaceCookie(cookie *http.Cookie) {
	for i, existing := range c.token.Cookies {
		if existing.Name == cookie.Name {
			c.token.Cookies[i] = cookie
			return
		}
	}
	c.token.Cookies = append(c.token.Cookies, cookie)
}

// parseSetCookieHeader safely extracts cookies from a raw Set-Cookie header line.
func parseSetCookieHeader(v string) []*http.Cookie {
	resp := &http.Response{Header: http.Header{"Set-Cookie": []string{v}}}
	return resp.Cookies()
}
