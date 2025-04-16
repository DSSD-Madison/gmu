package services

import (
	"context"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	log       logger.Logger
	dbQuerier *db.Queries
}

func NewLoginService(log logger.Logger, db *db.Queries) *LoginService {
	serviceLogger := log.With("service", "Login")
	return &LoginService{
		log:       serviceLogger,
		dbQuerier: db,
	}
}

func (lm *LoginService) ValidateUser(ctx context.Context, username string) (db.User, error) {
	lm.log.DebugContext(ctx, "Validating User", "username", username)
	user, err := lm.dbQuerier.GetUserByUsername(ctx, username)
	if err != nil {
		lm.log.Error("Login error: user not found", "username", username, "error", err)
		return db.User{}, err
	}
	lm.log.DebugContext(ctx, "Validated User", "username", username)
	return user, nil
}

func (lm *LoginService) ValidatePassword(user db.User, password string) error {
	lm.log.Debug("Validating Password", "username", user.Username)
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		lm.log.Error("Login error: incorrect password", "error", err)
		return err
	}
	lm.log.Debug("Validated Password", "username", user.Username)
	return nil
}

func (lm *LoginService) CreateSession(c echo.Context, user db.User) (*sessions.Session, error) {
	session, err := middleware.Store.Get(c.Request(), "session")
	if err != nil {
		lm.log.Error("Failed to get session info", "user", user.Username)
		return nil, err
	}
	session.Values["authenticated"] = true
	session.Values["username"] = user.Username
	session.Values["is_master"] = user.IsMaster
	session.Save(c.Request(), c.Response())
	return session, nil
}

func (lm *LoginService) Logout(c echo.Context) error {
	session, err := middleware.Store.Get(c.Request(), "session")
	if err != nil {
		lm.log.ErrorContext(c.Request().Context(), "Failed to log out", "error", err)
		return err
	}
	session.Values["authenticated"] = false
	session.Values["is_master"] = false
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		lm.log.ErrorContext(c.Request().Context(), "Failed to save session", "error", err)
		return err
	}
	return nil
}
