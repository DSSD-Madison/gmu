package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/labstack/echo/v4"
)

func RegisterUserManagementRoutes(e *echo.Echo, userManagementHandler *handlers.UserManagementHandler, sessionManager services.SessionManager) {
	e.GET("/admin/users", userManagementHandler.ManageUsersPage, sessionManager.RequireAuth)
	e.POST("/admin/users", userManagementHandler.CreateNewUser, sessionManager.RequireAuth)
	e.POST("/admin/users/delete", userManagementHandler.DeleteUser, sessionManager.RequireAuth)
}
