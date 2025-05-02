package container

import (
	"github.com/DSSD-Madison/gmu/pkg/aws/bedrock"
	"github.com/DSSD-Madison/gmu/pkg/aws/kendra"
	"github.com/DSSD-Madison/gmu/pkg/aws/s3"
	"github.com/DSSD-Madison/gmu/pkg/core/config"
	"github.com/DSSD-Madison/gmu/pkg/core/logger"
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/services"
)

type Container struct {
	Config *config.Config
	Logger logger.Logger

	BedrockClient bedrock.Client
	KendraClient  kendra.Client
	S3Client      s3.S3Client

	SearchService     services.Searcher
	SuggestionService services.Suggester

	HomeHandler   *handlers.HomeHandler
	SearchHandler *handlers.SearchHandler
}
