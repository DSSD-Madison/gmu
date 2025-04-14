package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	redirect := c.FormValue("redirect")

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}

	// Try to fetch user
	user, err := h.db.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		fmt.Println("Login error: user not found:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.LoginFormPartial("Invalid credentials", csrf, redirect))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect))
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		fmt.Println("Login error: incorrect password:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.LoginFormPartial("Invalid credentials", csrf, redirect))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect))
	}

	// Success: create session
	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["authenticated"] = true
	session.Values["username"] = user.Username
	session.Values["is_master"] = user.IsMaster
	session.Save(c.Request(), c.Response())

	// Secure redirect
	if redirect == "" || !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		redirect = "/"
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", redirect)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusSeeOther, redirect)
}


func (h *Handler) LoginPage(c echo.Context) error {
	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}
	redirect := c.QueryParam("redirect")
	return web.Render(c, http.StatusOK, components.LoginPage("", csrf, redirect))
}

