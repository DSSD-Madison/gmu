package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

const (
	sessionKeyAuthenticated = "authenticated"
	sessionKeyUserID        = "user_id"
	sessionKeyIsMaster      = "is_master"
)

type GorillaSessionManager struct {
	store       sessions.Store
	sessionName string
	log         logger.Logger
	db          *db.Queries
}

func NewGorillaSessionManager(store sessions.Store, sessionName string, log logger.Logger, dbClient *db.Queries) (*GorillaSessionManager, error) {
	if store == nil {
		return nil, errors.New("session store cannot be nil")
	}
	if sessionName == "" {
		return nil, errors.New("session name cannot be empty")
	}
	serviceLogger := log.With("Service", "Gorilla Session Manager")
	return &GorillaSessionManager{
		store:       store,
		sessionName: sessionName,
		log:         serviceLogger,
		db:          dbClient,
	}, nil
}

func (sm *GorillaSessionManager) getSession(r *http.Request) (*sessions.Session, error) {
	session, err := sm.store.Get(r, sm.sessionName)
	if err != nil {
		sm.log.WarnContext(r.Context(), "Error getting session", "error", err)
	}
	return session, err
}

func (sm *GorillaSessionManager) Create(c echo.Context, user db.User) error {
	session, _ := sm.getSession(c.Request())
	if session == nil {
		sm.log.ErrorContext(c.Request().Context(), "Failed to get or create session object", "user", user.Username)
		return errors.New("failed to initialize session")
	}

	session.Values[sessionKeyAuthenticated] = true
	session.Values[sessionKeyUserID] = user.ID.String()
	session.Values[sessionKeyIsMaster] = user.IsMaster

	err := session.Save(c.Request(), c.Response())
	if err != nil {
		sm.log.ErrorContext(c.Request().Context(), "Failed to save session after create", "error", err, "user", user.Username)
	}
	sm.log.InfoContext(c.Request().Context(), "Session Created", "user_id", user.ID, "username", user.Username)
	return nil
}

func (sm *GorillaSessionManager) Destroy(c echo.Context) error {
	session, _ := sm.getSession(c.Request())
	if session == nil {
		sm.log.ErrorContext(c.Request().Context(), "Failed to get or create session object for destroy")
		return errors.New("failed to initialize session for destroy")
	}

	session.Values[sessionKeyAuthenticated] = false
	delete(session.Values, sessionKeyUserID)
	delete(session.Values, sessionKeyIsMaster)
	session.Options.MaxAge = -1

	err := session.Save(c.Request(), c.Response())
	if err != nil {
		sm.log.ErrorContext(c.Request().Context(), "Failed to save session after destroy", "error", err)
	}
	return nil
}

func (sm *GorillaSessionManager) GetUserID(c echo.Context) (string, bool) {
	session, err := sm.getSession(c.Request())
	if err != nil || session == nil {
		return "", false
	}

	if authenticated, ok := session.Values[sessionKeyAuthenticated].(bool); !ok || !authenticated {
		return "", false
	}

	userID, ok := session.Values[sessionKeyUserID].(string)
	if !ok || userID == "" {
		return "", false
	}

	return userID, true
}

func (sm *GorillaSessionManager) IsAuthenticated(c echo.Context) bool {
	session, err := sm.getSession(c.Request())
	if err != nil || session == nil {
		return false
	}
	authenticated, ok := session.Values[sessionKeyAuthenticated].(bool)
	return ok && authenticated
}

func (sm *GorillaSessionManager) IsMaster(c echo.Context) bool {
	if !sm.IsAuthenticated(c) {
		return false
	}

	session, err := sm.getSession(c.Request())
	if err != nil || session == nil {
		return false
	}

	isMaster, ok := session.Values[sessionKeyIsMaster].(bool)
	return ok && isMaster
}

func (sm *GorillaSessionManager) redirectToLogin(c echo.Context) error {
	redirectTo := url.QueryEscape(c.Request().RequestURI)

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/login?redirect="+redirectTo)
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.Redirect(http.StatusSeeOther, "/login?redirect="+redirectTo)
}

// forceLogoutAndRedirect destroys the session and redirects the user to login.
func (sm *GorillaSessionManager) forceLogoutAndRedirect(c echo.Context) error {
	_ = sm.Destroy(c) 
	return sm.redirectToLogin(c)
}

func (sm *GorillaSessionManager) fetchUserByID(ctx context.Context, userID string) (*db.User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		sm.log.WarnContext(ctx, "Invalid UUID format for userID", "userID", userID)
		return nil, nil
	}

	row, err := sm.db.GetUserByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Manual conversion
	user := &db.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		IsMaster:     row.IsMaster,
		CreatedAt:    row.CreatedAt,
	}

	return user, nil
}

func (sm *GorillaSessionManager) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, _ := sm.getSession(c.Request())
		auth, ok := session.Values["authenticated"].(bool)
		sm.log.Info("requireAuth called", "auth", auth, "ok", ok)

		if !ok || !auth {
			sm.log.Info("requireAuth: Not authenticated or invalid session")
			return sm.redirectToLogin(c)
		}

		userID, ok := session.Values[sessionKeyUserID].(string)
		if !ok || userID == "" {
			sm.log.Warn("requireAuth: Missing or invalid user_id in session")
			return sm.forceLogoutAndRedirect(c)
		}

		ctx := c.Request().Context()

		user, err := sm.fetchUserByID(ctx, userID)
		if err != nil {
			sm.log.ErrorContext(ctx, "requireAuth: Error checking user existence in DB", "error", err)
			return sm.forceLogoutAndRedirect(c)
		}

		if user == nil {
			sm.log.WarnContext(ctx, "requireAuth: User ID no longer exists, forcing logout", "user_id", userID)
			return sm.forceLogoutAndRedirect(c)
		}

		return next(c)
	}
}

