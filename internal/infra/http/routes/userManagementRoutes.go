package routes

import (
	"github.com/DSSD-Madison/gmu/internal/application"
	"github.com/DSSD-Madison/gmu/internal/infra/http/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterUserManagementRoutes(e *echo.Echo, userManagementHandler *handlers.UserManagementHandler, sessionManager application.SessionManager) {
	e.GET("/admin/users", userManagementHandler.ManageUsersPage, sessionManager.RequireAuth)
	e.POST("/admin/users", userManagementHandler.CreateNewUser, sessionManager.RequireAuth)
	e.POST("/admin/users/delete", userManagementHandler.DeleteUser, sessionManager.RequireAuth)
}
