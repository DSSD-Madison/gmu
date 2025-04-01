package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) Home(c echo.Context) error {
	return web.Render(c, http.StatusOK, components.Home())
}
