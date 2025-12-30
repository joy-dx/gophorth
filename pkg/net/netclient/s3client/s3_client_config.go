package s3client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joy-dx/gophorth/pkg/net/netdto"
)

const NetClientS3Ref netdto.NetClientType = "net.client.s3"

// S3ClientConfig defines the static properties for an S3 client instance.
type S3ClientConfig struct {
	Region         string
	Credentials    aws.CredentialsProvider
	Middlewares    []netdto.Middleware
	ForcePathStyle bool
	Endpoint       string // optional custom endpoint
}

// S3RequestConfig defines the structure of an S3 request operation.
type S3RequestConfig struct {
	Operation string // "get", "put", "delete", "list"
	Bucket    string
	Key       string

	// Optional depending on operation
	Body        []byte
	Prefix      string
	ContentType string
	ExtraOpts   map[string]interface{}
	Headers     map[string]string
}

// Default config helpers
func DefaultS3ClientConfig(region string) S3ClientConfig {
	return S3ClientConfig{Region: region, Middlewares: []netdto.Middleware{}}
}

func (c *S3ClientConfig) WithMiddleware(m ...netdto.Middleware) *S3ClientConfig {
	c.Middlewares = append(c.Middlewares, m...)
	return c
}

func (c *S3RequestConfig) Ref() netdto.NetClientType {
	return NetClientS3Ref
}
