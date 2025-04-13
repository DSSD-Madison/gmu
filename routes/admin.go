package routes

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/pkg/db"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) ManageUsersPage(c echo.Context) error {
	if !middleware.IsMaster(c) {
		return c.String(http.StatusForbidden, "Access denied")
	}

	users, err := h.db.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	csrf := c.Get("csrf").(string)
	return web.Render(c, http.StatusOK, components.ManageUsersForm(csrf, "", users))
}


func (h *Handler) CreateNewUser(c echo.Context) error {
	if !middleware.IsMaster(c) {
		return c.String(http.StatusForbidden, "Access denied")
	}

	username := strings.TrimSpace(c.FormValue("username"))
	password := c.FormValue("password")
	confirm := c.FormValue("confirm_password")
	csrf := c.Get("csrf").(string)

	users, _ := h.db.ListUsers(c.Request().Context()) // Get users up front for reuse

	if password != confirm {
		return web.Render(c, http.StatusBadRequest, components.ManageUsersForm(csrf, "Passwords do not match", users))
	}

	// Optional: check if user already exists
	_, err := h.db.GetUserByUsername(c.Request().Context(), username)
	if err == nil {
		return web.Render(c, http.StatusConflict, components.ManageUsersForm(csrf, "User already exists", users))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return web.Render(c, http.StatusInternalServerError, components.ManageUsersForm(csrf, "Error hashing password", users))
	}

	err = h.db.CreateUser(c.Request().Context(), db.CreateUserParams{
		Username:     username,
		PasswordHash: string(hash),
		IsMaster:     false,
	})
	if err != nil {
		return web.Render(c, http.StatusInternalServerError, components.ManageUsersForm(csrf, "Failed to create user", users))
	}

	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
