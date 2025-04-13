package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
)

func (h *Handler) Logout(c echo.Context) error {
	session, _ := middleware.Store.Get(c.Request(), "session")
	session.Values["authenticated"] = false
	session.Values["is_master"] = false
	session.Options.MaxAge = -1
	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/")
}
