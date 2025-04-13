package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	csrf := c.Get("csrf").(string)

	// Lookup user
	user, err := h.db.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		return web.Render(c, http.StatusUnauthorized, components.LoginFormWithError("Invalid credentials", csrf))
	}

	// Compare password hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return web.Render(c, http.StatusUnauthorized, components.LoginFormWithError("Invalid credentials", csrf))
	}

	// Create session
	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["authenticated"] = true
	session.Values["username"] = user.Username
	session.Values["is_master"] = user.IsMaster
	session.Save(c.Request(), c.Response())

	redirect := c.QueryParam("redirect")
	if redirect == "" {
		redirect = "/"
	}
	return c.Redirect(http.StatusSeeOther, "/upload")
}


func (h *Handler) LoginPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	return web.Render(c, http.StatusOK, components.LoginForm(csrf))
}

