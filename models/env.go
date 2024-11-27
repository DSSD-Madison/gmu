package models

import (
	"os"

	"github.com/joho/godotenv"
)

var _ = godotenv.Load()

var accessKey = os.Getenv("ACCESS_KEY")
var secretKey = os.Getenv("SECRET_KEY")
var region = os.Getenv("REGION")
var indexId = os.Getenv("INDEX_ID")
