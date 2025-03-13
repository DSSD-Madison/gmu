// Environment provides access to the current program's environment
package environment

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joho/godotenv"
)

type environment struct {
	accessKey string
	secretKey string
	region    string
	indexId   string
}

func (e environment) AccessKey() string {
	return e.accessKey
}

func (e environment) SecretKey() string {
	return e.secretKey
}

func (e environment) Region() string {
	return e.region
}

func (e environment) IndexId() string {
	return e.indexId
}

func Environment() environment {
	return env
}

func AccessKey() string {
	return env.accessKey
}

func SecretKey() string {
	return env.secretKey
}

func Region() string {
	return env.region
}

func IndexId() string {
	return env.indexId
}

var env = func() environment {
	godotenv.Load()
	return environment{
		accessKey: os.Getenv("ACCESS_KEY"),
		secretKey: os.Getenv("SECRET_KEY"),
		region:    os.Getenv("REGION"),
		indexId:   os.Getenv("INDEX_ID"),
	}
}()

var creds = aws.Credentials{
	AccessKeyID:     AccessKey(),
	SecretAccessKey: SecretKey(),
}

type provider struct {
	Credentials aws.Credentials
}

func (p provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return p.Credentials, nil
}

func Provider() provider {
	return prov
}

func Credentials() aws.Credentials {
	return creds
}

var prov = provider{
	Credentials: creds,
}
