package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/joy-dx/gophorth/pkg/net/netclient/httpclient"
	"github.com/joy-dx/gophorth/pkg/net/netclient/s3client"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

func LoggingMiddleware(logger func(msg string)) netdto.Middleware {
	return func(ctx context.Context, inCfg *netdto.RequestConfig) error {
		switch cfg := inCfg.ReqConfig.(type) {
		case *httpclient.HTTPRequestConfig:
			logger(fmt.Sprintf("[HTTP]  %s %s", cfg.Method, cfg.URL))
		case *s3client.S3RequestConfig:
			logger(fmt.Sprintf("[S3]    %s s3://%s/%s", strings.ToUpper(cfg.Operation), cfg.Bucket, cfg.Key))
		default:
			logger("[NET] unknown request type")
			return errors.New("unsupported request type for logging")
		}
		return nil
	}
}
