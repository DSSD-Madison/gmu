package config

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joho/godotenv"
)

type Provider struct {
	Credentials aws.Credentials
}

func (p Provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return p.Credentials, nil
}

type Config struct {
	Mode     string
	LogLevel string

	AWS struct {
		Region           string
		Credentials      Provider
		KendraIndexID    string
		BedrockModelID   string
		S3BucketName     string
		KeywordsFilePath string
		RetryMaxAttempts int
	}

	Database struct {
		Host     string
		User     string
		Name     string
		Password string
	}

	Session struct {
		Secret   string
		MaxAge   int
		Secure   bool
		SameSite http.SameSite
	}
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

	config := &Config{
		Mode:     lookupEnv("MODE", "dev"),
		LogLevel: lookupEnv("LOG_LEVEL", "info"),
	}

	// Set AWS config properties
	config.AWS.Region = os.Getenv("REGION")
	config.AWS.Credentials = creds
	config.AWS.KendraIndexID = os.Getenv("INDEX_ID")
	config.AWS.BedrockModelID = os.Getenv("MODEL_ID")
	config.AWS.KeywordsFilePath = os.Getenv("KEYWORDS_FILE_PATH")
	config.AWS.S3BucketName = "manually-uploaded-bep"
	config.AWS.RetryMaxAttempts = 10

	// Set Database config properties
	config.Database.Host = os.Getenv("DB_HOST")
	config.Database.User = os.Getenv("DB_USER")
	config.Database.Name = os.Getenv("DB_NAME")
	config.Database.Password = os.Getenv("DB_PASSWORD")

	// Set session config properties
	config.Session.MaxAge = 86400 * 7
	config.Session.Secure = config.Mode == "prod"
	config.Session.Secret = os.Getenv("SESSION_SECRET_KEY")
	if config.Session.Secret == "" {
		panic("cannot have empty session secret. Please set SESSION_SECRET_KEY environment variable.")
	}
	config.Session.SameSite = http.SameSiteLaxMode

	return config, nil
}

func lookupEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
