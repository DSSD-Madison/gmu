package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	us.log.DebugContext(ctx, "Attempting to get user",
		"username", username)
	user, err := us.dbQuerier.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			us.log.InfoContext(ctx, "User not found",
				"username", username, "error", err)
			return db.User{}, fmt.Errorf("user %s not found: %w", username, err)
		}
		us.log.ErrorContext(ctx, "Database error getting user by username",
			"username", username,
			"error", err)
		return db.User{}, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	us.log.DebugContext(ctx, "Successfully retrived user",
		"username", username,
		"user_id", user.ID)
	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context) (db.User, error) {
	us.log.WarnContext(ctx, "CreateUser not implemented")
	return db.User{}, errors.New("user creation not implemented")
}

func (us *UserService) UpdateUser(ctx context.Context) (db.User, error) {
	us.log.WarnContext(ctx, "UpdateUser not implemented")
	return db.User{}, errors.New("user update not implemented")
}

func (us *UserService) DeleteUser(ctx context.Context) (db.User, error) {
	us.log.WarnContext(ctx, "DeleteUser not implemented")
	return db.User{}, errors.New("user deletion not implemented")
}
