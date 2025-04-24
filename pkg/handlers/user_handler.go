package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type UserManagementHandler struct {
	log logger.Logger
	db  *db.Queries
}

func NewUserManagementHandler(log logger.Logger, db *db.Queries) *UserManagementHandler {
	handlerLogger := log.With("handler", "UserManagement")
	return &UserManagementHandler{
		log: handlerLogger,
		db:  db,
	}
}

func RegisterUserManagementRoutes(e echo.Echo, umh *UserManagementHandler) {

}

func (uh *UserManagementHandler) ManageUsersPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	isAuthorized, isMaster := middleware.GetSessionFlags(c)

	if !isMaster {
		return c.String(http.StatusForbidden, "Access denied")
	}

	users, err := uh.db.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return web.Render(c, http.StatusOK, components.ManageUsersForm(csrf, "", users, isAuthorized, isMaster))
}

func (uh *UserManagementHandler) CreateNewUser(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	if !isMaster {
		return c.String(http.StatusForbidden, "Access denied")
	}

	username := strings.TrimSpace(c.FormValue("username"))
	password := c.FormValue("password")
	confirm := c.FormValue("confirm_password")

	users, _ := uh.db.ListUsers(c.Request().Context()) // Get users up front for reuse

	if password != confirm {
		return web.Render(c, http.StatusBadRequest, components.ManageUsersForm(csrf, "Passwords do not match", users, isAuthorized, isMaster))
	}

	// Optional: check if user already exists
	_, err := uh.db.GetUserByUsername(c.Request().Context(), username)
	if err == nil {
		return web.Render(c, http.StatusConflict, components.ManageUsersForm(csrf, "User already exists", users, isAuthorized, isMaster))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return web.Render(c, http.StatusInternalServerError, components.ManageUsersForm(csrf, "Error hashing password", users, isAuthorized, isMaster))
	}

	err = uh.db.CreateUser(c.Request().Context(), db.CreateUserParams{
		Username:     username,
		PasswordHash: string(hash),
		IsMaster:     false,
	})
	if err != nil {
		return web.Render(c, http.StatusInternalServerError, components.ManageUsersForm(csrf, "Failed to create user", users, isAuthorized, isMaster))
	}

	return c.Redirect(http.StatusSeeOther, "/admin/users")
}

func (uh *UserManagementHandler) DeleteUser(c echo.Context) error {
	_, isMaster := middleware.GetSessionFlags(c)
	if !isMaster {
		return c.String(http.StatusForbidden, "Access denied")
	}

	username := c.FormValue("username")
	if username == "" {
		return c.String(http.StatusBadRequest, "Username required")
	}

	// Prevent deleting master users just in case
	user, err := uh.db.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		return c.String(http.StatusNotFound, "User not found")
	}
	if user.IsMaster {
		return c.String(http.StatusForbidden, "Cannot delete admin users")
	}

	err = uh.db.DeleteUserByUsername(c.Request().Context(), username)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
