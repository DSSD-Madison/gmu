package models

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var creds = aws.Credentials{
	AccessKeyID:     accessKey,
	SecretAccessKey: secretKey,
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
