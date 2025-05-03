package application

import (
	"context"

	"github.com/DSSD-Madison/gmu/internal/infra/aws/s3"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type FilemanagerService struct {
	log      logger.Logger
	s3Client *s3.S3Client
}

func NewFilemanagerService(log logger.Logger, s3Client *s3.S3Client) *FilemanagerService {
	return &FilemanagerService{log: log, s3Client: s3Client}
}

func (fs *FilemanagerService) UploadFile(ctx context.Context, key string, data []byte, contentType string) error {
	return fs.s3Client.Upload(ctx, key, data, contentType)
}

func (fs *FilemanagerService) DeleteFile(ctx context.Context, key string, bucket string) error {
	return fs.s3Client.Delete(ctx, bucket, key)
}
