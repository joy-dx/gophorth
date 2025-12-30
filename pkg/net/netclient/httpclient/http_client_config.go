package httpclient

import (
	"net/http"
	"time"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"golang.org/x/oauth2"
)

type HTTPRequestConfig struct {
	Method string `json:"method" yaml:"method"`
	URL    string
	Body   map[string]interface{} `json:"body" yaml:"body"`
	// BodyType application/json, application/x-www-form-urlencoded
	BodyType string            `json:"body_type" yaml:"body_type"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
}

func DefaultHTTPRequestConfig() HTTPRequestConfig {
	return HTTPRequestConfig{
		Method:   http.MethodGet,
		Body:     map[string]interface{}{},
		BodyType: "application/json",
		Headers:  make(map[string]string),
	}
}

func (c *HTTPRequestConfig) Ref() netdto.NetClientType {
	return NetClientHTTPRef
}

func (c *HTTPRequestConfig) WithMethod(method string) *HTTPRequestConfig {
	c.Method = method
	return c
}
func (c *HTTPRequestConfig) WithBody(body map[string]interface{}) *HTTPRequestConfig {
	c.Body = body
	return c
}
func (c *HTTPRequestConfig) WithHeaders(headers map[string]string) *HTTPRequestConfig {
	c.Headers = headers
	return c
}
func (c *HTTPRequestConfig) WithURL(url string) *HTTPRequestConfig {
	c.URL = url
	return c
}

type HTTPClientConfig struct {
	AuthProvider  netdto.AuthProvider
	OAuthSource   oauth2.TokenSource
	RefreshBuffer time.Duration
	Middlewares   []netdto.Middleware
}

func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		RefreshBuffer: 30 * time.Second,
		Middlewares:   make([]netdto.Middleware, 0),
	}
}

// WithRefreshBuffer sets the early-refresh buffer.
func (c *HTTPClientConfig) WithAuthProvider(provider netdto.AuthProvider) *HTTPClientConfig {
	c.AuthProvider = provider
	return c
}
func (c *HTTPClientConfig) WithOAuthSource(tokenSource oauth2.TokenSource) *HTTPClientConfig {
	c.OAuthSource = tokenSource
	return c
}
func (c *HTTPClientConfig) WithRefreshBuffer(d time.Duration) *HTTPClientConfig {
	c.RefreshBuffer = d
	return c
}

// UseMiddleware appends pre-request Middlewares.
func (c *HTTPClientConfig) WithMiddleware(m ...netdto.Middleware) *HTTPClientConfig {
	c.Middlewares = append(c.Middlewares, m...)
	return c
}
