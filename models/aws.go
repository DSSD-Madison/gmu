package models

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
)

var creds = aws.Credentials{
	AccessKeyID:     access_key,
	SecretAccessKey: secret_key,
}

type Provider struct {
	Credentials aws.Credentials
}

func (p Provider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return p.Credentials, nil
}

var prov = Provider{
	Credentials: creds,
}
