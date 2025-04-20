package services

import (
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticationService struct {
	log       logger.Logger
	dbQuerier *db.Queries
}

func NewLoginService(log logger.Logger, db *db.Queries) *AuthenticationService {
	serviceLogger := log.With("service", "Login")
	return &AuthenticationService{
		log:       serviceLogger,
		dbQuerier: db,
	}
}

func (as *AuthenticationService) ValidateLogin(user db.User, password string) error {
	as.log.Debug("Validating Password", "username", user.Username)
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		as.log.Error("Login error: incorrect password", "error", err)
		return err
	}
	as.log.Debug("Validated Password", "username", user.Username)
	return nil
}
