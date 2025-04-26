package services

import (
	"context"
	"net/url"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/labstack/echo/v4"
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
	HandleLogin(ctx context.Context, ip string, username string, password string) (db.User, error)
}

type FileManager interface {
	UploadFile()
	GetMetadata()
	UpdateMetadata()
}

type BedrockManager interface {
	ExtractPDFMetadata(ctx context.Context, pdfBytes []byte) (*awskendra.ExtractedMetadata, error)
}

type SessionManager interface {
	Create(c echo.Context, user db.User) error
	Destroy(c echo.Context) error
	GetUserID(c echo.Context) (string, bool)
	IsAuthenticated(c echo.Context) bool
	IsMaster(c echo.Context) bool
	RequireAuth(next echo.HandlerFunc) echo.HandlerFunc
}
