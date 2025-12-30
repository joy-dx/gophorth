package s3client

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
	"github.com/joy-dx/gophorth/pkg/net/netutils"
)

func (c *S3Client) ProcessRequest(ctx context.Context, reqCfg *netdto.RequestConfig) (netdto.Response, error) {
	cfg, ok := reqCfg.ReqConfig.(*S3RequestConfig)
	if !ok {
		return netdto.Response{}, fmt.Errorf("problem casting to s3requestconfig")
	}

	// Apply middlewares
	for _, mw := range c.cfg.Middlewares {
		if err := mw(ctx, reqCfg); err != nil {
			return netdto.Response{}, fmt.Errorf("middleware aborted: %w", err)
		}
	}

	var (
		resp netdto.Response
		err  error
	)

	switch cfg.Operation {
	case "get":
		resp, err = c.doGet(ctx, cfg)
	case "put":
		resp, err = c.doPut(ctx, cfg)
	case "delete":
		resp, err = c.doDelete(ctx, cfg)
	case "list":
		resp, err = c.doList(ctx, cfg)
	default:
		err = fmt.Errorf("unsupported s3 operation: %s", cfg.Operation)
	}

	return resp, err
}

func (c *S3Client) doGet(ctx context.Context, cfg *S3RequestConfig) (netdto.Response, error) {
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(cfg.Key),
	})
	if err != nil {
		return netdto.Response{}, fmt.Errorf("s3 get object: %w", err)
	}
	defer out.Body.Close()
	data, err := io.ReadAll(out.Body)
	if err != nil {
		return netdto.Response{}, fmt.Errorf("read s3 object: %w", err)
	}
	return netdto.Response{
		StatusCode: 200,
		Body:       data,
		Headers:    netutils.MapToHeader(out.Metadata),
	}, nil
}

func (c *S3Client) doPut(ctx context.Context, cfg *S3RequestConfig) (netdto.Response, error) {
	bodyReader := bytes.NewReader(cfg.Body)
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(cfg.Key),
		Body:        bodyReader,
		ContentType: aws.String(cfg.ContentType),
	})
	if err != nil {
		return netdto.Response{}, fmt.Errorf("s3 put object: %w", err)
	}
	return netdto.Response{StatusCode: 200}, nil
}

func (c *S3Client) doDelete(ctx context.Context, cfg *S3RequestConfig) (netdto.Response, error) {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(cfg.Key),
	})
	if err != nil {
		return netdto.Response{}, fmt.Errorf("s3 delete object: %w", err)
	}
	return netdto.Response{StatusCode: 200}, nil
}

func (c *S3Client) doList(ctx context.Context, cfg *S3RequestConfig) (netdto.Response, error) {
	out, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.Bucket),
		Prefix: aws.String(cfg.Prefix),
	})
	if err != nil {
		return netdto.Response{}, fmt.Errorf("s3 list objects: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	for _, obj := range out.Contents {
		fmt.Fprintf(buf, "%s\n", aws.ToString(obj.Key))
	}

	return netdto.Response{
		StatusCode: 200,
		Body:       buf.Bytes(),
	}, nil
}
