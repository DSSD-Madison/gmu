package application

import (
	"context"

	"github.com/DSSD-Madison/gmu/internal/infra/aws/bedrock"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type BedrockService struct {
	log           logger.Logger
	bedrockClient bedrock.BedrockClient
}

func NewBedrockService(log logger.Logger, client bedrock.BedrockClient) *BedrockService {
	serviceLogger := log.With("service", "Bedrock")
	return &BedrockService{
		log:           serviceLogger,
		bedrockClient: client,
	}
}

func (b *BedrockService) ExtractPDFMetadata(ctx context.Context, pdfBytes []byte) (*bedrock.ExtractedMetadata, error) {
	metadata, err := b.bedrockClient.ProcessDocAndExtractMetadata(ctx, pdfBytes)
	if err != nil {
		b.log.ErrorContext(ctx, "failed to extract metadata from pdf", "error", err)
		return nil, err
	}

	return metadata, nil
}
