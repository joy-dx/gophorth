package netdto

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/joy-dx/gophorth/pkg/delay"
)

type TransferNotification struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Message     string `json:"message,omitempty" yaml:"message,omitempty"`
	// Status MetaType of message
	Status TransferStatus `json:"status" yaml:"status"`
	// Percentage completion status as a percentage
	Percentage float64 `json:"percentage" yaml:"percentage"`
	// TotalSize length content in bytes. The value -1 indicates that the length is unknown
	TotalSize int64 `json:"total_size,omitempty" yaml:"total_size,omitempty"`
	// Downloaded downloaded body length in bytes
	Downloaded int64 `json:"downloaded,omitempty" yaml:"downloaded,omitempty"`
}

type NetState struct {
	ExtraHeaders             ExtraHeaders  `json:"net_extra_headers,omitempty" yaml:"net_extra_headers,omitempty"`
	RequestTimeout           time.Duration `json:"net_request_timeout,omitempty" yaml:"net_request_timeout,omitempty"`
	UserAgent                string        `json:"net_user_agent,omitempty" yaml:"net_user_agent,omitempty"`
	BlacklistDomains         []string      `json:"net_blacklist_domains,omitempty" yaml:"net_blacklist_domains,omitempty"`
	WhitelistDomains         []string      `json:"net_whitelist_domains,omitempty" yaml:"net_whitelist_domains,omitempty"`
	DownloadCallbackInterval time.Duration `json:"net_download_callback_interval,omitempty" yaml:"net_download_callback_interval,omitempty"`
	// PreferCurlDownloads Instead of using imroc/req for downloads, prefer to use curl found on $PATH if available
	PreferCurlDownloads bool                            `json:"prefer_curl_downloads,omitempty" yaml:"net_prefer_curl_downloads,omitempty"`
	TransfersStatus     map[string]TransferNotification `json:"net_transfers_status,omitempty" yaml:"net_transfers_status,omitempty"`
}

// Download File
type DownloadFileConfig struct {
	// Blocking Determine whether or not program execution should wait
	Blocking bool
	Checksum string
	URL      string
	// DestinationFolder Used if path not set appending
	DestinationFolder string
	OutputFileName    string
	SkipAllowedPaths  bool
}

func DefaultDownloadFileConfig() DownloadFileConfig {
	return DownloadFileConfig{
		Blocking: true,
	}
}

func (c *DownloadFileConfig) WithBlocking(blocking bool) *DownloadFileConfig {
	c.Blocking = blocking
	return c
}

func (c *DownloadFileConfig) WithChecksum(checksum string) *DownloadFileConfig {
	c.Checksum = checksum
	return c
}

func (c *DownloadFileConfig) WithURL(url string) *DownloadFileConfig {
	c.URL = url
	return c
}

func (c *DownloadFileConfig) WithDestinationFolder(path string) *DownloadFileConfig {
	c.DestinationFolder = path
	return c
}

func (c *DownloadFileConfig) WithOutputFilename(name string) *DownloadFileConfig {
	c.OutputFileName = name
	return c
}

func (c *DownloadFileConfig) WithSkipAllowedPaths(truthy bool) *DownloadFileConfig {
	c.SkipAllowedPaths = truthy
	return c
}

type RequestConfig struct {
	// ClientRef Determines which http agent to use
	ClientRef string      `json:"client_ref" yaml:"client_ref"`
	ReqConfig interface{} `json:"req_config" yaml:"req_config"`
	// ResponseObject Used for casting result to
	ResponseObject interface{}      `json:"response_object" yaml:"response_object"`
	Timeout        time.Duration    `json:"timeout" yaml:"timeout"`
	MaxRetries     int              `json:"max_retries" yaml:"max_retries"`
	Delay          delay.RetryDelay `json:"-" yaml:"-"`
	TaskName       string           `json:"task_name" yaml:"task_name"`
}

func DefaultRequestConfig() RequestConfig {
	return RequestConfig{
		ClientRef:  NET_DEFAULT_CLIENT_REF,
		Timeout:    20 * time.Second,
		MaxRetries: 3,
		Delay:      delay.ExponentialBackoff{},
	}
}

func (c *RequestConfig) WithClientRef(ref string) *RequestConfig {
	c.ClientRef = ref
	return c
}

func (c *RequestConfig) WithReqConfig(cfg interface{}) *RequestConfig {
	c.ReqConfig = cfg
	return c
}

func (c *RequestConfig) WithResponseObject(object interface{}) *RequestConfig {
	c.ResponseObject = object
	return c
}

func (c *RequestConfig) WithTimeout(duration time.Duration) *RequestConfig {
	c.Timeout = duration
	return c
}

func (c *RequestConfig) WithMaxRetries(count int) *RequestConfig {
	c.MaxRetries = count
	return c
}

func (c *RequestConfig) WithDelay(delay delay.RetryDelay) *RequestConfig {
	c.Delay = delay
	return c
}

func (c *RequestConfig) WithTaskName(name string) *RequestConfig {
	c.TaskName = name
	return c
}

type Response struct {
	StatusCode int
	Headers    http.Header
	// As well as casting to ResponseObject if set, return as byes
	Body []byte
}

// TokenInfo represents active credential or session data.
// It supports both header-based tokens and cookie-based sessions.
type TokenInfo struct {
	// Authorization token, e.g. "Bearer abc123" or "Basic Zm9vOmJhcg=="
	AccessToken string
	// TokenType is inferred if not provided (default "Bearer").
	TokenType string
	// Expiry time. Optional — empty for cookie-only sessions.
	Expiry  time.Time
	Cookies []*http.Cookie
}

// IsExpired returns true if the token is close to or past expiry.
func (t *TokenInfo) IsExpired(buffer time.Duration) bool {
	if t.AccessToken == "" && len(t.Cookies) == 0 {
		return true
	}
	if t.Expiry.IsZero() {
		// Sessions with no expiry are considered indefinitely valid
		return false
	}
	return time.Now().After(t.Expiry.Add(-buffer))
}

// ExtraHeaders type is a comma seperated key=value string defined for use with Viper appconfig parsing
type ExtraHeaders map[string]string

func (e ExtraHeaders) String() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}

// Set Value should be a comma seperated key=value string
func (e ExtraHeaders) Set(s string) error {
	// First split by comma
	for _, header := range strings.Split(s, ",") {
		// Then split by = sign
		headerList := strings.Split(header, "=")
		// Set map
		e[headerList[0]] = headerList[1]
	}
	return nil
}

func (e ExtraHeaders) Type() string {
	return "ExtraHeaders"
}
