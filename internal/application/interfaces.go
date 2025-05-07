package application

import (
	"context"
	"net/url"

	"github.com/DSSD-Madison/gmu/internal/domain/search"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/bedrock"
	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/labstack/echo/v4"
)

type Searcher interface {
	SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (search.Results, error)
}

type Suggester interface {
	GetSuggestions(ctx context.Context, query string) (search.Suggestions, error)
}

type UserManager interface {
	CreateUser(ctx context.Context) (db.User, error)
	GetUser(ctx context.Context, username string) (db.User, error)
	UpdateUser(ctx context.Context) (db.User, error)
	DeleteUser(ctx context.Context) (db.User, error)
}

type AuthenticationManager interface {
	HandleLogin(ctx context.Context, ip string, username string, password string) (db.User, error)
}

type FileManager interface {
	UploadFile()
	GetMetadata()
	UpdateMetadata()
}

type BedrockManager interface {
	ExtractPDFMetadata(ctx context.Context, pdfBytes []byte) (*bedrock.ExtractedMetadata, error)
}

type SessionManager interface {
	Create(c echo.Context, user db.User) error
	Destroy(c echo.Context) error
	GetUserID(c echo.Context) (string, bool)
	IsAuthenticated(c echo.Context) bool
	IsMaster(c echo.Context) bool
	RequireAuth(next echo.HandlerFunc) echo.HandlerFunc
}
