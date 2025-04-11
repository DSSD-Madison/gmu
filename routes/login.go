package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	c.Logger().Infof("Login attempt with username='%s', password='%s'", username, password)

	if username == "admin" && password == "password" {
		session, _ := middleware.Store.Get(c.Request(), "session")
		session.Values["authenticated"] = true

		err := session.Save(c.Request(), c.Response())
		if err != nil {
			c.Logger().Errorf("Failed to save session: %v", err)
			return err
		}
		c.Logger().Info("Session saved successfully")

		redirect := c.QueryParam("redirect")
		if redirect == "" {
			redirect = "/"
		}

		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", redirect)
			c.Logger().Info("HX-Redirect sent")
			return c.NoContent(http.StatusOK)
		}

		c.Logger().Infof("Redirecting to: %s", redirect)
		return c.Redirect(http.StatusSeeOther, redirect)
	}

	c.Logger().Info("Invalid login")
	return web.Render(c, http.StatusUnauthorized, components.LoginFormWithError("Invalid credentials"))
}

func (h *Handler) LoginPage(c echo.Context) error {
	return web.Render(c, http.StatusOK, components.LoginForm())
}

