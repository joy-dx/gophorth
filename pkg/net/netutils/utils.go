package netutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func PrepareBody(body map[string]interface{}, bodyType string) ([]byte, string, error) {
	if body == nil {
		return nil, "", nil
	}

	switch strings.ToLower(bodyType) {
	case "application/json":
		buf, err := json.Marshal(body)
		return buf, "application/json", err
	case "application/x-www-form-urlencoded":
		vals := url.Values{}
		for k, v := range body {
			vals.Set(k, fmt.Sprintf("%v", v))
		}
		return []byte(vals.Encode()), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", fmt.Errorf("unsupported body_type: %s", bodyType)
	}
}

func IsTemporaryErr(err error) bool {
	// You could enhance this to check for net.Error timeouts etc.
	var netErr interface{ Temporary() bool }
	if errors.As(err, &netErr) {
		return netErr.Temporary()
	}
	// consider all network-level issues as transient
	return true
}

func MapToHeader(m map[string]string) http.Header {
	h := make(http.Header)
	for k, v := range m {
		h.Set(k, v)
	}
	return h
}
