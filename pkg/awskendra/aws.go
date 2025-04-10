package awskendra

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joho/godotenv"
)

type Config struct {
	Credentials Provider
	Region string
	IndexID string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	creds := Provider{aws.Credentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}}

	return &Config{
		Credentials: creds,
		Region: os.Getenv("REGION"),
		IndexID: os.Getenv("INDEX_ID"),
	}, nil
}

type Provider struct {
	Credentials aws.Credentials
}

func (p Provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return p.Credentials, nil
}

