package httpclient

import (
	"context"
	"errors"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

// StaticHeaderMiddleware injects static headers into every request.
func StaticHeaderMiddleware(headers map[string]string) netdto.Middleware {
	return func(ctx context.Context, inCfg *netdto.RequestConfig) error {
		cfg, castOk := inCfg.ReqConfig.(*HTTPRequestConfig)
		if !castOk {
			return errors.New("problem casting to httprequestconfig")
		}
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string)
		}
		for k, v := range headers {
			cfg.Headers[k] = v
		}
		inCfg.ReqConfig = cfg
		return nil
	}
}
