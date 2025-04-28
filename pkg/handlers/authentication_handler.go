package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type AuthenticationHandler struct {
	log                   logger.Logger
	userManager           services.UserManager
	authenticationManager services.AuthenticationManager
}

func NewAuthenticationHandler(log logger.Logger, userManager services.UserManager, authenticationManager services.AuthenticationManager) *AuthenticationHandler {
	handlerLogger := log.With("Handler", "Authentication")
	return &AuthenticationHandler{
		log:                   handlerLogger,
		userManager:           userManager,
		authenticationManager: authenticationManager,
	}
}

func (ah *AuthenticationHandler) RegisterAuthenticationRoutes(e *echo.Echo) {
	// Login Route
	e.GET("/login", ah.LoginPage)
	e.POST("/login", ah.Login)

	// Logout Route
	e.GET("/logout", ah.Logout) // for dev testing, remove when nav bar added
	e.POST("/logout", ah.Logout)
}

func (ah *AuthenticationHandler) CreateSession(c echo.Context, user db.User) (*sessions.Session, error) {
	session, err := middleware.Store.Get(c.Request(), "session")
	if err != nil {
		ah.log.Error("Failed to get session info", "user", user.Username)
		return nil, err
	}
	session.Values["authenticated"] = true
	session.Values["username"] = user.Username
	session.Values["is_master"] = user.IsMaster
	if err := session.Save(c.Request(), c.Response()); err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to save session", "error", err)
	}
	return session, nil
}

func (ah *AuthenticationHandler) LogoutSession(c echo.Context) error {
	session, err := middleware.Store.Get(c.Request(), "session")
	if err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to log out", "error", err)
		return err
	}
	session.Values["authenticated"] = false
	session.Values["is_master"] = false
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to save session", "error", err)
		return err
	}
	return nil
}

func (ah *AuthenticationHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	redirect := c.FormValue("redirect")

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}

	isAuthorized, isMaster := middleware.GetSessionFlags(c)

	user, err := ah.userManager.GetUser(c.Request().Context(), username)
	if err != nil {
		fmt.Println("Login error: user not found:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.ErrorMessage("Invalid credentials"))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect, isAuthorized, isMaster))
	}

	err = ah.authenticationManager.ValidateLogin(user, password)
	if err != nil {
		fmt.Println("Login error: incorrect password:", err)
		if c.Request().Header.Get("HX-Request") == "true" {
			return web.Render(c, http.StatusOK, components.ErrorMessage("Invalid credentials"))
		}
		return web.Render(c, http.StatusOK, components.LoginPage("Invalid credentials", csrf, redirect, isAuthorized, isMaster))
	}

	// Success: create session
	_, err = ah.CreateSession(c, user)
	if err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to create session", "error", err)
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

func (ah *AuthenticationHandler) LoginPage(c echo.Context) error {
	csrf, ok := c.Get("csrf").(string)
	if !ok {
		csrf = ""
	}
	redirect := c.QueryParam("redirect")
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	return web.Render(c, http.StatusOK, components.LoginPage("", csrf, redirect, isAuthorized, isMaster))
}

func (ah *AuthenticationHandler) Logout(c echo.Context) error {
	err := ah.LogoutSession(c)
	if err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to log out", "error", err)
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/")
}
