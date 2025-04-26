package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterAuthenticationRoutes(e *echo.Echo, authenticationHandler *handlers.AuthenticationHandler) {
	// Login Route
	e.GET("/login", authenticationHandler.LoginPage)
	e.POST("/login", authenticationHandler.Login)

	// Logout Route
	e.GET("/logout", authenticationHandler.Logout)
}
