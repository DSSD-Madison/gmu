// Package config defines structs and functions used for loading application-level configuration.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config stores application level configuration information.
type Config struct {
	// Mode represents the applications current runtime
	// environment such as Production or Development
	Mode string

	// LogLevel sets the application logger's
	// output specificity.
	LogLevel string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		Mode:     lookupEnv("MODE", "dev"),
		LogLevel: lookupEnv("LOG_LEVEL", "info"),
	}, nil
}

func lookupEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
