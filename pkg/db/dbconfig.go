package db

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost string
	DBUser string
	DBName string
	DBPassword string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		DBHost: os.Getenv("PROD_HOST"),
		DBUser: os.Getenv("PROD_USER"),
		DBName: os.Getenv("PROD_DB"),
		DBPassword: os.Getenv("PROD_PASSWORD"),
	}, nil
}
