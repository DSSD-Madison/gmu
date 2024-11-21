package models

import (
	"github.com/joho/godotenv"
	"os"
)

var gdenverr = godotenv.Load()

var access_key = os.Getenv("ACCESS_KEY")
var secret_key = os.Getenv("SECRET_KEY")
var region = os.Getenv("REGION")
var index_id = os.Getenv("INDEX_ID")

type env struct {
	Access_Key string
	Secret_Key string
	Region     string
	Index_Id   string
}

var environment = env{
	Access_Key: access_key,
	Secret_Key: secret_key,
	Region:     region,
	Index_Id:   index_id,
}

func GetEnv() env {
	return environment
}
