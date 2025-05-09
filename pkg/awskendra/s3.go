package awskendra

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client provides methods to interact with the AWS S3 service.
type S3Client struct {
	client *s3.Client
	config Config
}

// NewS3Client constructs an S3Client with the given config.
func NewS3Client(cfg Config) (*S3Client, error) {

	opts := aws.Config{
		Region:           cfg.Region,
		Credentials:      cfg.Credentials,
		RetryMaxAttempts: cfg.RetryMaxAttempts,
	}

	client := s3.NewFromConfig(opts)
	s3c := &S3Client{
		client: client,
		config: cfg,
	}

	return s3c, nil
}

func (s *S3Client) Upload(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             &s.config.BucketName,
		Key:                &key,
		Body:               bytes.NewReader(body),
		ContentType:        &contentType,
		ContentDisposition: aws.String("inline"),
	})
	return err
}

func (s *S3Client) Delete(ctx context.Context, bucket string, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}
