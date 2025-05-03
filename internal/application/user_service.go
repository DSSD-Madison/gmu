package application

import (
	"context"

	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type UserService struct {
	log       logger.Logger
	dbQuerier *db.Queries
}

func NewUserService(log logger.Logger, db *db.Queries) *UserService {
	serviceLogger := log.With("service", "Login")
	return &UserService{
		log:       serviceLogger,
		dbQuerier: db,
	}
}

func (us *UserService) GetUser(ctx context.Context, username string) (db.User, error) {
	us.log.DebugContext(ctx, "Getting user", "username", username)
	user, err := us.dbQuerier.GetUserByUsername(ctx, username)
	if err != nil {
		us.log.Error("Login error: user not found", "username", username, "error", err)
		return db.User{}, err
	}
	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context) (db.User, error) {
	return db.User{}, nil
}

func (us *UserService) UpdateUser(ctx context.Context) (db.User, error) {
	return db.User{}, nil
}

func (us *UserService) DeleteUser(ctx context.Context) (db.User, error) {
	return db.User{}, nil
}
