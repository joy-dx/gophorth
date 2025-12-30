package s3client

import (
	"context"

	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

// StaticS3MetaMiddleware adds default metadata to each S3 put operation.
func StaticS3MetaMiddleware(meta map[string]string) netdto.Middleware {
	return func(ctx context.Context, inCfg *netdto.RequestConfig) error {
		cfg, ok := inCfg.ReqConfig.(*S3RequestConfig)
		if !ok {
			return nil // not an S3 request
		}
		if cfg.ExtraOpts == nil {
			cfg.ExtraOpts = map[string]interface{}{}
		}
		for k, v := range meta {
			cfg.ExtraOpts[k] = v
		}
		return nil
	}
}
