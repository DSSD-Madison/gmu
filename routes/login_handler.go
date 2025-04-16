package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type LoginHandler struct {
	log     logger.Logger
	manager services.LoginManager
}

func NewLoginHandler(log logger.Logger, manager services.LoginManager) *LoginHandler {
	handlerLogger := log.With("Handler", "Login")
	return &LoginHandler{
		log:     handlerLogger,
		manager: manager,
	}
}

func (lh *LoginHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	redirect := c.FormValue("redirect")

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}

	isAuthorized, isMaster := middleware.GetSessionFlags(c)

	user, err := lh.manager.ValidateUser(c.Request().Context(), username)
	if err != nil {
		fmt.Println("Login error: user not found:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.ErrorMessage("Invalid credentials"))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect, isAuthorized, isMaster))
	}

	err = lh.manager.ValidatePassword(user, password)
	if err != nil {
		fmt.Println("Login error: incorrect password:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.ErrorMessage("Invalid credentials"))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect, isAuthorized, isMaster))
	}

	// Success: create session
	_, err = lh.manager.CreateSession(c, user)
	if err != nil {
		lh.log.ErrorContext(c.Request().Context(), "Failed to create session", "error", err)
		return err
	}

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

func (lh *LoginHandler) LoginPage(c echo.Context) error {
	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}
	redirect := c.QueryParam("redirect")
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	return web.Render(c, http.StatusOK, components.LoginPage("", csrf, redirect, isAuthorized, isMaster))
}

func (lh *LoginHandler) Logout(c echo.Context) error {
	err := lh.manager.Logout(c)
	if err != nil {
		lh.log.ErrorContext(c.Request().Context(), "Failed to log out", "error", err)
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/")
}
