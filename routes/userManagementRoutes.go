package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterUserManagementRoutes(e *echo.Echo, userManagementHandler *handlers.UserManagementHandler) {
	e.GET("/admin/users", userManagementHandler.ManageUsersPage, middleware.RequireAuth)
	e.POST("/admin/users", userManagementHandler.CreateNewUser, middleware.RequireAuth)
	e.POST("/admin/users/delete", userManagementHandler.DeleteUser, middleware.RequireAuth)
}
