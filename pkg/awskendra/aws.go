package awskendra

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joho/godotenv"
)

// LoadConfig loads the configuration for KendraClient. If successful, the returned config can be used to configure a Kendra Client
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
		Credentials:      creds,
		Region:           os.Getenv("REGION"),
		IndexID:          os.Getenv("INDEX_ID"),
		ModelID:          os.Getenv("MODEL_ID"),
		KeywordsFilePath: os.Getenv("KEYWORDS_FILE_PATH"),
		RetryMaxAttempts: 10,
	}, nil
}

// Provider implements the Provider interface provided by aws.
type Provider struct {
	Credentials aws.Credentials
}

// Retrieve returns the aws Credentials held by the Provider struct.
func (p Provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return p.Credentials, nil
}
