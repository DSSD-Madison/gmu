package services

import (
	"context"
	"net/url"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
)

type Searcher interface {
	SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (awskendra.KendraResults, error)
}

type Suggester interface {
	GetSuggestions(ctx context.Context, query string) (awskendra.KendraSuggestions, error)
}

type UserManager interface {
	CreateUser(ctx context.Context) (db.User, error)
	GetUser(ctx context.Context, username string) (db.User, error)
	UpdateUser(ctx context.Context) (db.User, error)
	DeleteUser(ctx context.Context) (db.User, error)
}

type AuthenticationManager interface {
	ValidateLogin(user db.User, password string) error
}

type FileManager interface {
	UploadFile()
	GetMetadata()
	UpdateMetadata()
}

type BedrockManager interface {
	ExtractPDFMetadata(ctx context.Context, pdfBytes []byte) (*awskendra.ExtractedMetadata, error)
}
