package internal

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode string
	LogLevel string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		Mode: lookupEnv("MODE", "dev"),
		LogLevel: lookupEnv("LOG_LEVEL", "info"),
	}, nil
}

func lookupEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

