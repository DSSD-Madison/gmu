package services

import (
	"context"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type BedrockService struct {
	log           logger.Logger
	bedrockClient awskendra.BedrockClient
}

func NewBedrockService(log logger.Logger, client awskendra.BedrockClient) *BedrockService {
	serviceLogger := log.With("service", "Bedrock")
	return &BedrockService{
		log:           serviceLogger,
		bedrockClient: client,
	}
}

func (b *BedrockService) ExtractPDFMetadata(ctx context.Context, pdfBytes []byte) (*awskendra.ExtractedMetadata, error) {
	metadata, err := b.bedrockClient.ProcessDocAndExtractMetadata(ctx, pdfBytes)
	if err != nil {
		b.log.ErrorContext(ctx, "failed to extract metadata from pdf", "error", err)
		return nil, err
	}

	return metadata, nil
}
