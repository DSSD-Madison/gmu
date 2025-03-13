package models

import (
	"context"

	"github.com/DSSD-Madison/gmu/models/environment"
	"github.com/aws/aws-sdk-go-v2/aws"
)

var creds = aws.Credentials{
	AccessKeyID:     environment.AccessKey(),
	SecretAccessKey: environment.SecretKey(),
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
