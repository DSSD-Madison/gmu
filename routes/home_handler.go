package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type HomeHandler struct {
	log logger.Logger
}

func NewHomeHandler(log logger.Logger) *HomeHandler {
	return &HomeHandler{log: log}
}

func (h *HomeHandler) Home(c echo.Context) error {
	h.log.InfoContext(c.Request().Context(), "Rendering home page")
	return web.Render(c, http.StatusOK, components.Home())
}
