package services

import (
	"context"
	"github.com/DSSD-Madison/gmu/pkg/awskendra"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type FilemanagerService struct {
	log      logger.Logger
	s3Client *awskendra.S3Client
}

func NewFilemanagerService(log logger.Logger, s3Client *awskendra.S3Client) *FilemanagerService {
	return &FilemanagerService{log: log, s3Client: s3Client}
}

func (fs *FilemanagerService) UploadFile(ctx context.Context, key string, data []byte) error {
	return fs.s3Client.Upload(ctx, key, data)
}

func (fs *FilemanagerService) DeleteFile(ctx context.Context, key string, bucket string) error {
	return fs.s3Client.Delete(ctx, bucket, key)
}
