package s3client

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

type S3Client struct {
	NetClient netdto.NetClient
	cfg       *S3ClientConfig
	client    *s3.Client
	mu        sync.RWMutex
}

func NewS3Client(ref string, cfg *S3ClientConfig) (*S3Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(cfg.Credentials),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &S3Client{
		cfg:    cfg,
		client: client,
		NetClient: netdto.NetClient{
			Name:        "S3 Client",
			Ref:         ref,
			ClientType:  NetClientS3Ref,
			Description: "Performs basic S3 operations (get, put, list, delete)",
		},
	}, nil
}

func (c *S3Client) Ref() string {
	return c.NetClient.Ref
}

func (c *S3Client) Type() netdto.NetClientType {
	return NetClientS3Ref
}
