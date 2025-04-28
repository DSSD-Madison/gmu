package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type AuthenticationHandler struct {
	log                   logger.Logger
	authenticationManager services.AuthenticationManager
	sessionManager        services.SessionManager
}

func NewAuthenticationHandler(log logger.Logger, sessionManager services.SessionManager, authenticationManager services.AuthenticationManager) *AuthenticationHandler {
	handlerLogger := log.With("Handler", "Authentication")
	return &AuthenticationHandler{
		log:                   handlerLogger,
		sessionManager:        sessionManager,
		authenticationManager: authenticationManager,
	}
}

func (ah *AuthenticationHandler) RegisterAuthenticationRoutes(e *echo.Echo) {
	// Login Route
	e.GET("/login", ah.LoginPage)
	e.POST("/login", ah.Login)

	e.POST("/logout", ah.Logout)
}

// Login TODO: Remove checks for HX-Header
func (ah *AuthenticationHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	redirect := c.FormValue("redirect")
	ctx := c.Request().Context()
	ip := c.RealIP()

	csrf, _ := c.Get("csrf").(string)

	user, err := ah.authenticationManager.HandleLogin(ctx, ip, username, password)
	if err != nil {
		ah.log.WarnContext(ctx, "Login failed", "error", err, "username", username, "ip", ip)

		userMessage := "Invalid Credentials"
		httpStatus := http.StatusUnauthorized

		if errors.Is(err, services.ErrRateLimited) {
			userMessage = "Too many attempts. Please try again later."
			httpStatus = http.StatusTooManyRequests
		} else if !errors.Is(err, services.ErrInvalidCredentials) {
			ah.log.ErrorContext(ctx, "Unexpected login error", "error", err, "user", username, "ip", ip)
			userMessage = "An unexpected error occurred."
			httpStatus = http.StatusInternalServerError
		}

		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().WriteHeader(httpStatus)
			ah.log.Info("HX-Request is true")
			return web.Render(c, httpStatus, components.ErrorMessage(userMessage))
		}

		isAuthorized := ah.sessionManager.IsAuthenticated(c)
		isMaster := ah.sessionManager.IsMaster(c)

		ah.log.Info("Rendering login page")
		return web.Render(c, http.StatusOK, components.LoginPage(userMessage, csrf, redirect, isAuthorized, isMaster))
	}

	ah.log.InfoContext(ctx, "Login successful, creating session", "username", user.Username)

	// Success: create session
	err = ah.sessionManager.Create(c, user)
	if err != nil {
		ah.log.ErrorContext(c.Request().Context(), "Failed to create session", "error", err)
		return web.Render(c, http.StatusInternalServerError, components.ErrorMessage("Login succeeded but failed to create session. Please try again."))
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
	csrf, _ := c.Get("csrf").(string)
	redirect := c.QueryParam("redirect")

	isAuthorized := ah.sessionManager.IsAuthenticated(c)
	isMaster := ah.sessionManager.IsMaster(c)
	return web.Render(c, http.StatusOK, components.LoginPage("", csrf, redirect, isAuthorized, isMaster))
}

func (ah *AuthenticationHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	err := ah.sessionManager.Destroy(c)
	if err != nil {
		ah.log.ErrorContext(ctx, "Error destroying session via manager", "error", err)
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/login")
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}
